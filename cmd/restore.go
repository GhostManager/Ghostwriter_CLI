package cmd

import (
	"fmt"
	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore <db backup filename> <media backup filename>",
	Short: "Restores the specified PostgreSQL database and media file backups",
	Long: `Restores the specified PostgreSQL database and media file backups stored it in the production_postgres_data_backups
Docker volume. Use the --list flag to list current backup files. Provide the full filenames of the files you want
to restore.

WARNING: Restoring cannot be undone! Your database will be overwritten and media files will be replaced.

This command runs PostgreSQL's dropdb and createdb commands to drop the current database and then recreate it using
the specified backup file. Backup files are gunzipped SQL files from pg_dump.

The media files are restored by wiping the media directory and then copying the files from the backup to the media volume.`,
	Args: cobra.ExactArgs(2),
	Run:  restoreDatabase,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

func restoreDatabase(cmd *cobra.Command, args []string) {
	dockerErr := internal.EvaluateDockerComposeStatus()
	if dockerErr == nil {
		fmt.Println("[!] Restoring this backup will overwrite the database and replace all media files (e.g., templates, evidence files).")
		c := internal.AskForConfirmation("Do you really want to restore this backup? This cannot be undone!")
		if c {
			if dev {
				internal.SetDevMode()
				fmt.Printf("[+] Restoring the `%s` backup file in the development environment...\n", args[0])
				internal.RunDockerComposeRestore("local.yml", args[0], args[1])
			} else {
				internal.SetProductionMode()
				fmt.Printf("[+] Restoring the `%s` backup file in the production environment...\n", args[0])
				internal.RunDockerComposeRestore("production.yml", args[0], args[1])

			}
		}
	}
}
