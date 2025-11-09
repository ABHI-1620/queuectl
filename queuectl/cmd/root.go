package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var dbPath string

var rootCmd = &cobra.Command{
	Use:   "queuectl",
	Short: "QueueCTL - A CLI-based background job queue system",
	Long:  `Manage background jobs, workers, retries, and DLQs via a simple CLI.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "queue.db", "Path to SQLite database file")
}
