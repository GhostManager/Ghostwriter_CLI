package cmd

import (
	"fmt"
	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
	"strings"
)

var mediaBackupFile string

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore <database backup filename>",
	Short: "Restores the specified PostgreSQL database backup and optionally media files",
	Long: `Restores the specified PostgreSQL database backup stored in the production_postgres_data_backups
Docker volume. Optionally restores media files using the --media flag. Use the backup command with --list flag 
to list current backup files. Provide the full filename of the file you want to restore.

WARNING: Restoring cannot be undone!

This command runs PostgreSQL's dropdb and createdb commands to drop the current database and then recreate it using
the specified backup file. Backup files are gunzipped SQL files from pg_dump. If a media backup is specified, it will
wipe the existing media files and restore them from the tar.gz archive.

Examples:
  # Restore only the database
  ghostwriter-cli restore backup_2023_05_23T15_54_19.sql.gz
  
  # Restore both database and media files
  ghostwriter-cli restore backup_2023_05_23T15_54_19.sql.gz --media media_backup_2023_05_23T15_54_19.tar.gz`,
	Args: cobra.ExactArgs(1),
	Run:  restoreDatabase,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().StringVar(&mediaBackupFile, "media", "", "Media backup filename to restore (optional)")
}

func restoreDatabase(cmd *cobra.Command, args []string) {
	dockerErr := internal.EvaluateDockerComposeStatus()
	if dockerErr == nil {
		confirmMsg := "Do you really want to restore this backup file? This cannot be undone!"
		if mediaBackupFile != "" {
			confirmMsg = "Do you really want to restore the database and media backups? This cannot be undone!"
		}
		c := internal.AskForConfirmation(confirmMsg)
		if c {
			if dev {
				internal.SetDevMode()
				fmt.Printf("[+] Restoring the `%s` database backup file in the development environment...\n", args[0])
				internal.RunDockerComposeRestore("local.yml", args[0])
				if mediaBackupFile != "" {
					if !strings.HasPrefix(mediaBackupFile, "media_backup_") {
						fmt.Println("[!] Warning: Media backup filename should start with 'media_backup_'")
					}
					fmt.Printf("[+] Restoring the `%s` media backup file in the development environment...\n", mediaBackupFile)
					internal.RunDockerComposeMediaRestore("local.yml", mediaBackupFile)
				}
			} else {
				internal.SetProductionMode()
				fmt.Printf("[+] Restoring the `%s` database backup file in the production environment...\n", args[0])
				internal.RunDockerComposeRestore("production.yml", args[0])
				if mediaBackupFile != "" {
					if !strings.HasPrefix(mediaBackupFile, "media_backup_") {
						fmt.Println("[!] Warning: Media backup filename should start with 'media_backup_'")
					}
					fmt.Printf("[+] Restoring the `%s` media backup file in the production environment...\n", mediaBackupFile)
					internal.RunDockerComposeMediaRestore("production.yml", mediaBackupFile)
				}
			}
		}
	}
}
