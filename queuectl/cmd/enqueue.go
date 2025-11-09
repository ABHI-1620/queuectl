package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ABHI-1620/queuectl/internal/db"
	"github.com/ABHI-1620/queuectl/internal/queue"
)

var (
	jobID      string
	jobCommand string
	jobRetries int
	jobPriority int
)

var enqueueCmd = &cobra.Command{
	Use:   "enqueue [job_json]",
	Short: "Enqueue a new job (accepts JSON or flags)",
	Long: `Add a new job to the queue.

Examples:
  queuectl enqueue --id job1 --cmd "echo Hello"
  queuectl enqueue '{"id":"job2","command":"sleep 5"}'
`,
	// ✅ allow arbitrary number of args (PowerShell might split)
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var in struct {
			ID         string `json:"id"`
			Command    string `json:"command"`
			MaxRetries int    `json:"max_retries,omitempty"`
			Priority   int    `json:"priority,omitempty"`
			RunAt      string `json:"run_at,omitempty"`
		}

		// ✅ Handle JSON mode only if first arg looks like '{'
		if len(args) > 0 && len(args[0]) > 0 && args[0][0] == '{' {
			if err := json.Unmarshal([]byte(args[0]), &in); err != nil {
				return fmt.Errorf("invalid job JSON: %w", err)
			}
		} else {
			// Use flags
			in.ID = jobID
			in.Command = jobCommand
			in.MaxRetries = jobRetries
			in.Priority = jobPriority
			// Merge any stray args from PowerShell
			if len(args) > 0 && in.Command == "" {
				in.Command = strings.Join(args, " ")
			}
		}

		if in.ID == "" || in.Command == "" {
			return fmt.Errorf("job id and command are required (use flags or JSON)")
		}

		runAt := time.Now()
		if in.RunAt != "" {
			if t, err := time.Parse(time.RFC3339, in.RunAt); err == nil {
				runAt = t
			}
		}

		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		job := queue.Job{
			ID:         in.ID,
			Command:    in.Command,
			MaxRetries: in.MaxRetries,
			Priority:   in.Priority,
			RunAt:      runAt,
		}

		if err := queue.Enqueue(context.Background(), database, job); err != nil {
			return err
		}

		fmt.Printf("✅ Job %s enqueued\n", job.ID)
		return nil
	},
}

func init() {
	enqueueCmd.Flags().StringVar(&jobID, "id", "", "Job ID")
	enqueueCmd.Flags().StringVar(&jobCommand, "cmd", "", "Command to run")
	enqueueCmd.Flags().IntVar(&jobRetries, "max-retries", 3, "Maximum retry count")
	enqueueCmd.Flags().IntVar(&jobPriority, "priority", 0, "Job priority")
	rootCmd.AddCommand(enqueueCmd)
}
