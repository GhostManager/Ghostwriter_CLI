package cmd

import (
	"fmt"
	"log"

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
	dockerInterface := docker.GetDockerInterface(mode)
	dockerInterface.Env.Save()

	if lst {
		listBackups(dockerInterface)
	} else {
		backup(dockerInterface)
	}
}

func listBackups(dockerInterface *docker.DockerInterface) {
	fmt.Printf("[+] Listing available PostgreSQL database backup files with %s...\n", dockerInterface.ComposeFile)
	err := dockerInterface.RunComposeCmd("run", "--rm", "postgres", "backups")
	if err != nil {
		log.Fatalf("Error trying to list backups files with %s: %v\n", dockerInterface.ComposeFile, err)
	}
}

func backup(dockerInterface *docker.DockerInterface) {
	fmt.Printf("[+] Backing up the PostgreSQL database with %s...\n", dockerInterface.ComposeFile)
	err := dockerInterface.RunComposeCmd("run", "--rm", "postgres", "backup")
	if err != nil {
		log.Fatalf("Error trying to back up the PostgreSQL database with %s: %v\n", dockerInterface.ComposeFile, err)
	}

	err = dockerInterface.BackupMediaFiles()
	if err != nil {
		log.Fatalf("Error trying to back up media files with %s: %v\n", dockerInterface.ComposeFile, err)
	}
}
