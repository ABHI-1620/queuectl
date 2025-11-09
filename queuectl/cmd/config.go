package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ABHI-1620/queuectl/internal/db"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage queue configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		_, err = database.Exec(`INSERT OR REPLACE INTO config(key, value, updated_at) VALUES (?, ?, datetime('now'))`, key, value)
		if err != nil {
			return err
		}

		fmt.Printf("✅ Config %s set to %s\n", key, value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "List current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		rows, _ := database.Query(`SELECT key, value FROM config`)
		defer rows.Close()
		fmt.Println("\n⚙️  Configuration:")
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			fmt.Printf("  %-20s : %s\n", k, v)
		}
		fmt.Println()
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	rootCmd.AddCommand(configCmd)
}
