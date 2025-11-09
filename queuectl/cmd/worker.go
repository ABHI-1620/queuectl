package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/ABHI-1620/queuectl/internal/db"
	"github.com/ABHI-1620/queuectl/internal/queue"
)

var workerCount int

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Manage background workers",
}

var workerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start background workers",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle SIGINT/SIGTERM
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sig
			fmt.Println("\n🛑 Stopping workers gracefully...")
			cancel()
		}()

		fmt.Printf("🚀 Starting %d workers...\n", workerCount)
		return queue.StartWorkers(ctx, database, workerCount)
	},
}

var workerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop workers (if managed via PID file)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Use Ctrl+C or system signal to stop workers gracefully.")
		return nil
	},
}

func init() {
	workerStartCmd.Flags().IntVar(&workerCount, "count", 1, "Number of workers to start")

	workerCmd.AddCommand(workerStartCmd)
	workerCmd.AddCommand(workerStopCmd)
	rootCmd.AddCommand(workerCmd)
}
