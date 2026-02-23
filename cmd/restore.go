package cmd

import (
	"fmt"
	"log"
	"strings"

	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
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
	dockerInterface := docker.GetDockerInterface(mode)

	// Validate that containers are running and match the current mode
	if err := dockerInterface.ValidateContainersRunning(); err != nil {
		log.Fatalf("%v\n", err)
	}

	confirmMsg := "Do you really want to restore this backup file? This cannot be undone!"
	if mediaBackupFile != "" {
		confirmMsg = "Do you really want to restore the database and media backups? This cannot be undone!"
	}
	c := internal.AskForConfirmation(confirmMsg)
	if !c {
		return
	}

	dockerInterface.Env.Save()

	fmt.Printf("[+] Restoring the `%s` database backup file...\n", args[0])
	restore(dockerInterface, args[0])
	if mediaBackupFile != "" {
		if !strings.HasPrefix(mediaBackupFile, "media_backup_") {
			fmt.Println("[!] Warning: Media backup filename should start with 'media_backup_'")
		}
		fmt.Printf("[+] Restoring the `%s` media backup file...\n", mediaBackupFile)
		mediaRestore(dockerInterface, mediaBackupFile)
	}
}

// RunDockerComposeRestore executes the "docker compose" command to restore a PostgreSQL database backup in the
// environment from the specified YAML file ("yaml" parameter).
func restore(dockerInterface *docker.DockerInterface, restore string) {
	fmt.Printf("[+] Restoring the PostgreSQL database backup file %s with %s...\n", restore, dockerInterface.ComposeFile)
	backupErr := dockerInterface.RunComposeCmd("run", "--rm", "postgres", "restore", restore)
	if backupErr != nil {
		log.Fatalf("Error trying to restore %s with %s: %v\n", restore, dockerInterface.ComposeFile, backupErr)
	}
}

func mediaRestore(dockerInterface *docker.DockerInterface, restore string) {
	// Determine the volume keys based on the environment
	var dataVolumeKey, backupVolumeKey string
	if dockerInterface.UseDevInfra {
		dataVolumeKey = "local_data"
		backupVolumeKey = "local_postgres_data_backups"
	} else {
		// Both production modes use the same volume keys
		dataVolumeKey = "production_data"
		backupVolumeKey = "production_postgres_data_backups"
	}

	// Get actual volume names from Docker Compose configuration
	dataVolume, err := dockerInterface.GetVolumeNameFromConfig(dataVolumeKey)
	if err != nil {
		log.Fatalf("Failed to get data volume name from compose config: %v\n", err)
	}

	backupVolume, err := dockerInterface.GetVolumeNameFromConfig(backupVolumeKey)
	if err != nil {
		log.Fatalf("Failed to get backup volume name from compose config: %v\n", err)
	}

	fmt.Printf("[+] Restoring media files from backup %s with %s...\n", restore, dockerInterface.ComposeFile)

	// First, clear the existing media files
	fmt.Println("[+] Clearing existing media files...")
	clearErr := dockerInterface.RunComposeCmd(
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/data", dataVolume),
		"postgres",
		"sh", "-c",
		"rm -rf /data/* /data/..?* /data/.[!.]*",
	)
	if clearErr != nil {
		log.Fatalf("Error trying to clear existing media files with %s: %v\n", dockerInterface.ComposeFile, clearErr)
	}

	// Extract the backup archive to the media volume
	fmt.Println("[+] Extracting media backup...")
	restoreErr := dockerInterface.RunComposeCmd(
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/data", dataVolume),
		"-v", fmt.Sprintf("%s:/backups:ro", backupVolume),
		"postgres",
		"sh", "-c",
		fmt.Sprintf("tar xzf /backups/%s -C /data", restore),
	)
	if restoreErr != nil {
		log.Fatalf("Error trying to restore media files from %s with %s: %v\n", restore, dockerInterface.ComposeFile, restoreErr)
	}
	fmt.Printf("[+] Media files restored from %s\n", restore)
}
