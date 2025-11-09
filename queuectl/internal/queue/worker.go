package queue

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/ABHI-1620/queuectl/internal/util"
)

func process(ctx context.Context, db *sql.DB, j *Job) error {
	out, exitCode, err := util.RunCommand(ctx, j.Command, 30*time.Second)
	if err == nil && exitCode == 0 {
		_, err := db.ExecContext(ctx, `UPDATE jobs SET state='completed', updated_at=datetime('now') WHERE id=?`, j.ID)
		fmt.Printf("✅ Job %s completed\n", j.ID)
		return err
	}

	// retry logic
	j.Attempts++
	if j.Attempts > j.MaxRetries {
		_, _ = db.ExecContext(ctx, `
		INSERT OR REPLACE INTO dlq(id, original_id, command, attempts, last_error)
		VALUES(?, ?, ?, ?, ?)`,
			j.ID+"-dlq", j.ID, j.Command, j.Attempts, err.Error())
		_, _ = db.ExecContext(ctx, `
		UPDATE jobs SET state='dead', updated_at=datetime('now'), last_error=? WHERE id=?`,
			err.Error(), j.ID)
		fmt.Printf("💀 Job %s moved to DLQ\n", j.ID)
		return nil
	}

	delay := math.Pow(float64(j.BaseBackoff), float64(j.Attempts))
	if delay > 1800 {
		delay = 1800
	}
	_, _ = db.ExecContext(ctx, `
	UPDATE jobs SET state='failed', attempts=?, next_run_at=datetime('now', ? || ' seconds'),
	last_error=?, updated_at=datetime('now') WHERE id=?`,
		j.Attempts, int(delay), fmt.Sprintf("exit=%d err=%v out=%s", exitCode, err, out), j.ID)
	fmt.Printf("⚠️  Job %s failed, retrying in %.0f sec\n", j.ID, delay)
	return nil
}

func StartWorkers(ctx context.Context, db *sql.DB, count int) error {
	hostname, _ := os.Hostname()
	for i := 0; i < count; i++ {
		go func(idx int) {
			workerID := fmt.Sprintf("%s-%d", hostname, idx)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					j, err := ClaimNext(ctx, db, workerID, 60)
					if err == ErrNoJob {
						time.Sleep(500 * time.Millisecond)
						continue
					}
					if err != nil {
						time.Sleep(200 * time.Millisecond)
						continue
					}
					process(ctx, db, j)
				}
			}
		}(i)
	}
	<-ctx.Done()
	return nil
}
