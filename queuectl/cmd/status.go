package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ABHI-1620/queuectl/internal/db"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show summary of all job states and active workers",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		showCounts(database)
		return nil
	},
}

func showCounts(database *sql.DB) {
	fmt.Println("\n📋 Job States:")
	rows, _ := database.Query(`SELECT state, COUNT(*) FROM jobs GROUP BY state`)
	defer rows.Close()
	for rows.Next() {
		var state string
		var count int
		rows.Scan(&state, &count)
		fmt.Printf("  %-10s : %d\n", state, count)
	}

	fmt.Println("\n🧑‍💻 Active Workers:")
	rows, _ = database.Query(`SELECT worker_id, COUNT(*) FROM jobs WHERE state='processing' GROUP BY worker_id`)
	defer rows.Close()
	for rows.Next() {
		var wid string
		var count int
		rows.Scan(&wid, &count)
		fmt.Printf("  %-10s : %d\n", wid, count)
	}
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
