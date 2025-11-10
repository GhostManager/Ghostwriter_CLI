package cmd

import (
	"fmt"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

var lst bool

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Creates a backup of the PostgreSQL database and media files",
	Long: `Creates a backup of the PostgreSQL database and media files, storing them in the "production_postgres_data_backups"
Docker volume as timestamped archives. The database backup is the result of PostgreSQL's pg_dump piped into gzip,
and the media backup is a tar.gz archive of the media files.

Use the --list flag to list current backup files.

Example files: 
  - backup_2023_05_23T15_54_19.sql.gz (database)
  - media_backup_2023_05_23T15_54_19.tar.gz (media files)`,
	Run: backupDatabase,
}

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().BoolVar(&lst, "list", false, "List the available backup files")
}

func backupDatabase(cmd *cobra.Command, args []string) {
	dockerErr := docker.EvaluateDockerComposeStatus()
	if dockerErr == nil {
		if dev {
			docker.SetDevMode()
			if lst {
				fmt.Println("[+] Getting a list of available backup files in the development environment")
				docker.RunDockerComposeBackups("local.yml")
			} else {
				fmt.Println("[+] Backing up the PostgreSQL database for the development environment")
				docker.RunDockerComposeBackup("local.yml")
				fmt.Println("[+] Backing up media files for the development environment")
				docker.RunDockerComposeMediaBackup("local.yml")
			}
		} else {
			docker.SetProductionMode()
			if lst {
				fmt.Println("[+] Getting a list of available backup files in the production environment")
				docker.RunDockerComposeBackups("production.yml")
			} else {
				fmt.Println("[+] Backing up the PostgreSQL database for the production environment")
				docker.RunDockerComposeBackup("production.yml")
				fmt.Println("[+] Backing up media files for the production environment")
				docker.RunDockerComposeMediaBackup("production.yml")
			}
		}
	}
}
