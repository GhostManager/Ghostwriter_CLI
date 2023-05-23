package cmd

import (
	"fmt"
	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restores the specified PostgreSQL database backup",
	Long: `Restores the specified PostgreSQL database backup stored it in the production_postgres_data_backups
Docker volume. Use the --list flag to list current backup files.

WARNING: Restoring cannot be undone!

This command runs PostgreSQL's dropdb and createdb commands to drop the current database and then recreate it using
the specified backup file. Backup files are gunzipped SQL files from pg_dump.`,
	Args: cobra.ExactArgs(1),
	Run:  restoreDatabase,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

func restoreDatabase(cmd *cobra.Command, args []string) {
	dockerErr := internal.EvaluateDockerComposeStatus()
	if dockerErr == nil {
		c := internal.AskForConfirmation("Do you really want to restore this backup file? This cannot be undone!")
		if c {
			if dev {
				internal.SetDevMode()
				fmt.Printf("[+] Restoring the `%s` backup file in the development environment...\n", args[0])
				internal.RunDockerComposeRestore("local.yml", args[0])
			} else {
				internal.SetProductionMode()
				fmt.Printf("[+] Restoring the `%s` backup file in the production environment...\n", args[0])
				internal.RunDockerComposeRestore("production.yml", args[0])

			}
		}
	}
}
