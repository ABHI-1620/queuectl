package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ABHI-1620/queuectl/internal/db"
)

var jobState string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs by state",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		rows, err := database.Query(`SELECT id, command, state, attempts, max_retries FROM jobs WHERE state=? ORDER BY created_at`, jobState)
		if err != nil {
			return err
		}
		defer rows.Close()

		fmt.Printf("\nJobs in state: %s\n", jobState)
		fmt.Println("----------------------------------------------------------")
		for rows.Next() {
			var id, cmdStr, state string
			var attempts, maxRetries int
			rows.Scan(&id, &cmdStr, &state, &attempts, &maxRetries)
			fmt.Printf("%-10s %-30s (%s) attempts=%d/%d\n", id, cmdStr, state, attempts, maxRetries)
		}
		fmt.Println()
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&jobState, "state", "pending", "Job state to list")
	rootCmd.AddCommand(listCmd)
}
