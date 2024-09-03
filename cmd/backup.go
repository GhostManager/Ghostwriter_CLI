package cmd

import (
	"fmt"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

var lst bool
var download bool
var yml string

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Creates a backup of the PostgreSQL database and media files",
	Long: `Creates a backup of the PostgreSQL database and media files and stores them in the
"production_postgres_data_backups" Docker volume as a timestamped gunzip. The database backup is the result of
PostgreSQL's pg_dump piped into gzip.

Use the --list flag to list current backup files.

Download the backup file(s) with the --download flag.

Example file: ghostwriter_2023_05_23T15_54_19.sql.gz`,
	Run: backupDatabase,
}

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().BoolVar(&lst, "list", false, "List the available backup file(s)")
	backupCmd.Flags().BoolVar(&download, "download", false, "Download the backup file(s)")
}

func backupDatabase(cmd *cobra.Command, args []string) {
	dockerErr := docker.EvaluateDockerComposeStatus()
	if dockerErr == nil {
		if dev {
			docker.SetDevMode()
			yml = "local.yml"
		} else {
			docker.SetProductionMode()
			yml = "production.yml"
		}
		if lst {
			fmt.Println("[+] Getting a list of available backup files...")
			docker.RunDockerListBackups(yml)
		} else if download {
			fmt.Println("[+] Downloading backup files...")
			docker.RunDockerDownloadBackups(yml)
		} else {
			fmt.Println("[+] Backing up the PostgreSQL database and media files...")
			docker.RunDockerComposeBackup(yml)
		}
	}
}
