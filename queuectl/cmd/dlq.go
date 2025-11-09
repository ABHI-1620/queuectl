package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ABHI-1620/queuectl/internal/db"
	"github.com/ABHI-1620/queuectl/internal/queue"
)

var dlqCmd = &cobra.Command{
	Use:   "dlq",
	Short: "Manage Dead Letter Queue jobs",
}

var dlqListCmd = &cobra.Command{
	Use:   "list",
	Short: "List DLQ jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		rows, _ := database.Query(`SELECT id, original_id, command, attempts, last_error FROM dlq`)
		defer rows.Close()

		fmt.Println("\n💀 Dead Letter Queue Jobs:")
		for rows.Next() {
			var id, originalID, cmdStr, errMsg string
			var attempts int
			rows.Scan(&id, &originalID, &cmdStr, &attempts, &errMsg)
			fmt.Printf("%-10s from %-10s [%d attempts] - %s\n", id, originalID, attempts, errMsg)
		}
		fmt.Println()
		return nil
	},
}

var dlqRetryCmd = &cobra.Command{
	Use:   "retry <job_id>",
	Short: "Retry a job from the DLQ",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jobID := args[0]
		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		var originalID, command string
		err = database.QueryRow(`SELECT original_id, command FROM dlq WHERE id=?`, jobID).Scan(&originalID, &command)
		if err != nil {
			return fmt.Errorf("DLQ job not found: %v", err)
		}

		job := queue.Job{
			ID:      originalID + "-retry",
			Command: command,
		}
		if err := queue.Enqueue(context.Background(), database, job); err != nil {
			return err
		}

		fmt.Printf("🔁 Retried DLQ job %s -> new job %s\n", jobID, job.ID)
		return nil
	},
}

func init() {
	dlqCmd.AddCommand(dlqListCmd)
	dlqCmd.AddCommand(dlqRetryCmd)
	rootCmd.AddCommand(dlqCmd)
}
