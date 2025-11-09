package queue

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Job struct {
	ID          string
	Command     string
	State       string
	Attempts    int
	MaxRetries  int
	BaseBackoff int
	Priority    int
	RunAt       time.Time
}

func Enqueue(ctx context.Context, db *sql.DB, j Job) error {
	if j.ID == "" {
		return errors.New("job id required")
	}
	if j.Command == "" {
		return errors.New("command required")
	}
	if j.MaxRetries == 0 {
		j.MaxRetries = 3
	}
	if j.BaseBackoff == 0 {
		j.BaseBackoff = 2
	}
	_, err := db.ExecContext(ctx, `
	INSERT OR REPLACE INTO jobs
	(id, command, state, attempts, max_retries, base_backoff, priority, run_at, created_at, updated_at)
	VALUES (?, ?, 'pending', 0, ?, ?, ?, datetime(?), datetime('now'), datetime('now'))
	`, j.ID, j.Command, j.MaxRetries, j.BaseBackoff, j.Priority, j.RunAt.Format(time.RFC3339))
	return err
}

var (
	ErrNoJob      = errors.New("no job available")
	ErrContention = errors.New("contention")
)

func ClaimNext(ctx context.Context, db *sql.DB, workerID string, leaseSec int) (*Job, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var id string
	row := tx.QueryRowContext(ctx, `
	SELECT id FROM jobs
	WHERE state IN ('pending','failed')
	AND COALESCE(next_run_at, run_at) <= datetime('now')
	ORDER BY priority ASC, created_at ASC
	LIMIT 1
	`)
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoJob
		}
		return nil, err
	}

	res, err := tx.ExecContext(ctx, `
	UPDATE jobs SET state='processing', worker_id=?, lease_expires_at=datetime('now', ? || ' seconds')
	WHERE id=? AND state IN ('pending','failed')
	`, workerID, leaseSec, id)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, ErrContention
	}

	var j Job
	err = tx.QueryRowContext(ctx, `
	SELECT id, command, attempts, max_retries, base_backoff
	FROM jobs WHERE id=?`, id).Scan(&j.ID, &j.Command, &j.Attempts, &j.MaxRetries, &j.BaseBackoff)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &j, nil
}
