package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrates user configuration files from current directory to the data directory to migrate a local installation to use the published images",
	Long: `Migrates custom configuration files and SSL certificates from the current working directory to the new
standardized data directory structure introduced in Ghostwriter CLI v6.3.0.

This command helps users transition from running Ghostwriter in their source repository to using the XDG data directory:
  - macOS: ~/Library/Application Support/ghostwriter/
  - Linux: ~/.local/share/ghostwriter/
  - Windows: %LOCALAPPDATA%/ghostwriter/

Files migrated:
  - SSL certificates: ssl/ghostwriter.crt, ssl/ghostwriter.key, ssl/dhparam.pem
  - Environment file: .env
  - Django settings: config/settings/production.d/* → settings/*

Docker volumes will also be migrated to preserve your database and media files:
  - PostgreSQL database (production_postgres_data)
  - Media uploads (production_data)
  - Backup archives (production_postgres_data_backups)
  - Static files are regenerated automatically (no migration needed)

The command will prompt for confirmation before overwriting any existing files in the destination.
Source files remain in place after migration (copy, not move).

Example:
  # Navigate to your old Ghostwriter directory and run:
  cd /path/to/old/ghostwriter
  ghostwriter-cli migrate`,
	Run: migrateFiles,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func migrateFiles(cmd *cobra.Command, args []string) {
	dockerInterface := internal.GetDockerInterface(mode)

	// Get source (CWD) and destination (data directory) paths
	sourcePath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v\n", err)
	}

	destPath := dockerInterface.Dir

	// Resolve absolute paths for comparison
	sourceAbs, err := filepath.Abs(sourcePath)
	if err != nil {
		log.Fatalf("Failed to resolve source path: %v\n", err)
	}

	destAbs, err := filepath.Abs(destPath)
	if err != nil {
		log.Fatalf("Failed to resolve destination path: %v\n", err)
	}

	// Check if source and destination are the same
	if sourceAbs == destAbs {
		log.Fatalf("Source and destination directories are the same (%s). Nothing to migrate.\n", sourceAbs)
	}

	// Warn if containers are running
	runningContainers := dockerInterface.GetRunning()
	if len(runningContainers) > 0 {
		fmt.Printf("[!] Warning: Found %d running Ghostwriter container(s).\n", len(runningContainers))
		if !internal.AskForConfirmation("It's recommended to stop containers before migrating. Continue anyway?") {
			fmt.Println("Migration cancelled. Consider running 'ghostwriter-cli down' first.")
			return
		}
	}

	fmt.Printf("[+] Migrating files from:\n    %s\n[+] To:\n    %s\n", sourceAbs, destAbs)

	// Track overall statistics
	totalMigrated := 0
	totalSkipped := 0
	totalFailed := 0
	var allErrors []error

	// Migrate SSL certificates
	fmt.Println("[+] Checking for SSL certificates...")
	sslMigrated, sslSkipped, sslErrors := migrateSSL(sourceAbs, destAbs)
	totalMigrated += sslMigrated
	totalSkipped += sslSkipped
	totalFailed += len(sslErrors)
	allErrors = append(allErrors, sslErrors...)

	// Migrate .env file
	fmt.Println("[+] Checking for .env file...")
	envMigrated, envError := migrateEnvFile(sourceAbs, destAbs)
	if envMigrated {
		totalMigrated++
	} else if envError != nil {
		totalFailed++
		allErrors = append(allErrors, envError)
	} else {
		totalSkipped++
	}

	// Migrate Django settings
	fmt.Println("[+] Checking for Django settings...")
	settingsMigrated, settingsSkipped, settingsErrors := migrateSettings(sourceAbs, destAbs)
	totalMigrated += settingsMigrated
	totalSkipped += settingsSkipped
	totalFailed += len(settingsErrors)
	allErrors = append(allErrors, settingsErrors...)

	// Migrate Docker volumes if old installation detected
	fmt.Println("[+] Checking for Docker volumes to migrate...")
	volumesMigrated, volumesSkipped, volumeErrors := migrateVolumes(dockerInterface, sourceAbs)
	if len(volumeErrors) > 0 {
		allErrors = append(allErrors, volumeErrors...)
		totalFailed += len(volumeErrors)
	}
	// Note: volumes are counted separately below

	// Print summary
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Migration Summary")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("  Files migrated:   %d\n", totalMigrated)
	fmt.Printf("  Files skipped:    %d\n", totalSkipped)
	fmt.Printf("  Files failed:     %d\n", totalFailed)
	if volumesMigrated > 0 || volumesSkipped > 0 {
		fmt.Printf("  Volumes migrated: %d\n", volumesMigrated)
		fmt.Printf("  Volumes skipped:  %d\n", volumesSkipped)
	}

	if len(allErrors) > 0 {
		fmt.Println("Errors encountered:")
		for i, err := range allErrors {
			fmt.Printf("  %d. %v\n", i+1, err)
		}
	}

	if totalMigrated > 0 {
		fmt.Println("[+] Migration complete!")
		fmt.Println("[+] Next steps:")
		fmt.Println("    1. Review migrated files in the data directory")
		fmt.Println("    2. Run 'ghostwriter-cli up' to start with the new configuration")
	} else if totalFailed == 0 {
		fmt.Println("[+] No files needed migration.")
	}
}

// migrateSSL migrates SSL certificate files from source to destination.
func migrateSSL(sourceBase, destBase string) (migrated, skipped int, errors []error) {
	sslFiles := []string{"ghostwriter.crt", "ghostwriter.key", "dhparam.pem"}

	// Check both "ssl" directories as possible sources
	sourceDirs := []string{
		filepath.Join(sourceBase, "ssl"),
	}

	var sourceSSLDir string
	for _, dir := range sourceDirs {
		if internal.DirExists(dir) {
			sourceSSLDir = dir
			break
		}
	}

	if sourceSSLDir == "" {
		fmt.Println("    No SSL directory found (checked ssl/)")
		return 0, 0, nil
	}

	fmt.Printf("    Found SSL directory: %s\n", filepath.Base(sourceSSLDir))
	destSSLDir := filepath.Join(destBase, "ssl")

	for _, filename := range sslFiles {
		sourcePath := filepath.Join(sourceSSLDir, filename)
		destPath := filepath.Join(destSSLDir, filename)

		if !internal.FileExists(sourcePath) {
			fmt.Printf("    ⊝ %s: not found\n", filename)
			skipped++
			continue
		}

		// Use 0600 for private keys, 0644 for others
		perm := os.FileMode(0644)
		if filename == "ghostwriter.key" {
			perm = 0600
		}

		wasMigrated, err := internal.MigrateFile(sourcePath, destPath, perm, true)
		if err != nil {
			fmt.Printf("    ✗ %s: failed (%v)\n", filename, err)
			errors = append(errors, fmt.Errorf("SSL file %s: %w", filename, err))
			continue
		}

		if wasMigrated {
			fmt.Printf("    ✓ %s: migrated\n", filename)
			migrated++
		} else {
			fmt.Printf("    ⊖ %s: skipped\n", filename)
			skipped++
		}
	}

	return
}

// migrateEnvFile migrates the .env file from source to destination.
func migrateEnvFile(sourceBase, destBase string) (bool, error) {
	sourcePath := filepath.Join(sourceBase, ".env")
	destPath := filepath.Join(destBase, ".env")

	if !internal.FileExists(sourcePath) {
		fmt.Println("    .env: not found")
		return false, nil
	}

	wasMigrated, err := internal.MigrateFile(sourcePath, destPath, 0600, true)
	if err != nil {
		fmt.Printf("    ✗ .env: failed (%v)\n", err)
		return false, fmt.Errorf(".env file: %w", err)
	}

	if wasMigrated {
		fmt.Println("    ✓ .env: migrated")
		return true, nil
	}

	fmt.Println("    ⊖ .env: skipped")
	return false, nil
}

// migrateSettings migrates Django settings files from the source to destination.
func migrateSettings(sourceBase, destBase string) (migrated, skipped int, errors []error) {
	sourceSettingsDir := filepath.Join(sourceBase, "config", "settings", "production.d")
	destSettingsDir := filepath.Join(destBase, "settings")

	if !internal.DirExists(sourceSettingsDir) {
		fmt.Println("    No Django settings directory found (config/settings/production.d/)")
		return 0, 0, nil
	}

	fmt.Printf("    Found settings directory: config/settings/production.d/\n")

	result, err := internal.MigrateDirectory(sourceSettingsDir, destSettingsDir, true)
	if err != nil {
		errors = append(errors, fmt.Errorf("settings directory: %w", err))
		fmt.Printf("    ✗ Failed to migrate settings: %v\n", err)
		return 0, 0, errors
	}

	if result.Migrated > 0 {
		fmt.Printf("    ✓ Migrated %d settings file(s)\n", result.Migrated)
	}
	if result.Skipped > 0 {
		fmt.Printf("    ⊖ Skipped %d file(s)\n", result.Skipped)
	}
	if result.Failed > 0 {
		fmt.Printf("    ✗ Failed %d file(s)\n", result.Failed)
	}

	return result.Migrated, result.Skipped, result.Errors
}

// migrateVolumes detects and migrates Docker volumes from an old installation to the new setup.
// This specifically migrates PRODUCTION volumes (not local dev volumes).
func migrateVolumes(dockerInterface *internal.DockerInterface, sourceDir string) (migrated, skipped int, errors []error) {
	// Check if an old production compose file exists in the source directory
	// Note: We look for production.yml specifically because we're migrating production volumes
	oldComposeFile := filepath.Join(sourceDir, "production.yml")
	if !internal.FileExists(oldComposeFile) {
		// Try alternative name (some old installations used docker-compose.yml)
		oldComposeFile = filepath.Join(sourceDir, "docker-compose.yml")
		if !internal.FileExists(oldComposeFile) {
			fmt.Println("    No old production compose file found - skipping volume migration")
			return 0, 0, nil
		}
	}

	fmt.Printf("    Found old compose file: %s\n", filepath.Base(oldComposeFile))

	// Define production volumes to migrate (not local dev volumes)
	// These match the volume keys used in production.yml
	volumesToMigrate := []string{
		"production_postgres_data",
		"production_data",
		"production_postgres_data_backups",
	}

	// Ensure containers are stopped
	if len(dockerInterface.GetRunning()) > 0 {
		fmt.Println("    Stopping containers before volume migration...")
		if err := dockerInterface.Down(nil); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop containers: %w", err))
			return 0, 0, errors
		}
		// Wait for containers to fully stop and release volumes
		fmt.Println("    Waiting for containers to stop...")
		for i := 0; i < 5; i++ {
			fmt.Print(".")
			if len(dockerInterface.GetRunning()) == 0 {
				break
			}
			time.Sleep(1 * time.Second)
		}
		fmt.Println()
	}

	// Create backup before migration
	if internal.AskForConfirmation("Create safety backup of the current contents of the destination volumes before volume migration?") {
		fmt.Println("    Creating backup (this may take a few minutes)...")
		// Temporarily start containers for backup
		if err := dockerInterface.Up(); err == nil {
			if err := dockerInterface.RunComposeCmd("run", "--rm", "postgres", "backup"); err != nil {
				fmt.Printf("    ⊖ Backup failed (continuing anyway): %v\n", err)
			} else {
				fmt.Println("    ✓ Backup created successfully")
			}
			// Stop containers again
			dockerInterface.Down(nil)
		} else {
			fmt.Printf("    [!] Warning: Failed to start containers for backup: %v\n", err)
			fmt.Println("    [!] Continuing without backup - you may want to create a manual backup")
		}
	}

	// Check if user wants to migrate volumes
	if !internal.AskForConfirmation("Migrate Docker volumes? This will copy database and media files to new volumes") {
		fmt.Println("    Skipping volume migration")
		return 0, len(volumesToMigrate), nil
	}

	// List old production volumes to verify they exist.
	// We intentionally scope this to the canonical legacy project name so we
	// only migrate well-known/official volume names.
	const legacyProductionPrefix = "ghostwriter_production"
	oldVolumes, err := dockerInterface.ListVolumes(legacyProductionPrefix)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to list old volumes: %w", err))
		return 0, 0, errors
	}

	if len(oldVolumes) == 0 {
		fmt.Println("    No old production volumes found - skipping volume migration")
		return 0, 0, nil
	}

	fmt.Printf("    Found %d old production volume(s) with prefix %q\n", len(oldVolumes), legacyProductionPrefix)

	volumeSourceMap := map[string]string{}
	for _, volumeKey := range volumesToMigrate {
		matchedVolume := findLegacyVolumeByKey(oldVolumes, volumeKey)
		if matchedVolume != "" {
			volumeSourceMap[volumeKey] = matchedVolume
		}
	}

	// Migrate each production volume
	for _, volumeKey := range volumesToMigrate {
		oldVolumeName := volumeSourceMap[volumeKey]

		if oldVolumeName == "" {
			fmt.Printf("    ⊝ %s: not found\n", volumeKey)
			skipped++
			continue
		}

		// Get the new volume name from the new compose file (in XDG data directory)
		// The dockerInterface already points to the new compose file
		newVolumeName, err := dockerInterface.GetVolumeNameFromConfig(volumeKey)
		if err != nil {
			fmt.Printf("    ✗ %s: failed to get new volume name (%v)\n", volumeKey, err)
			errors = append(errors, fmt.Errorf("volume %s: %w", volumeKey, err))
			continue
		}

		// Skip if volumes are the same
		if oldVolumeName == newVolumeName {
			fmt.Printf("    ⊖ %s: already using correct volume name\n", volumeKey)
			skipped++
			continue
		}

		// Copy the volume
		fmt.Printf("    Migrating %s...\n", volumeKey)
		if err := dockerInterface.CopyVolume(oldVolumeName, newVolumeName); err != nil {
			fmt.Printf("    ✗ %s: failed (%v)\n", volumeKey, err)
			errors = append(errors, fmt.Errorf("volume %s: %w", volumeKey, err))
			continue
		}

		// Verify the copy
		srcCount, destCount, err := dockerInterface.VerifyVolumeCopy(oldVolumeName, newVolumeName)
		if err != nil {
			fmt.Printf("    ⚠ %s: migrated but verification failed (%v)\n", volumeKey, err)
		} else if srcCount != destCount {
			fmt.Printf("    ⚠ %s: migrated but file counts differ (source: %d, dest: %d)\n", volumeKey, srcCount, destCount)
		} else {
			fmt.Printf("    ✓ %s: migrated (%d files)\n", volumeKey, destCount)
		}

		migrated++
	}

	// Offer to clean up old volumes
	if migrated > 0 {
		fmt.Println()
		if internal.AskForConfirmation("Delete old volumes to free disk space? (migrated data is preserved)") {
			for _, oldVolumeName := range volumeSourceMap {
				if err := dockerInterface.RunCmd("volume", "rm", oldVolumeName); err != nil {
					fmt.Printf("    ⊖ Failed to delete %s: %v\n", oldVolumeName, err)
				} else {
					fmt.Printf("    ✓ Deleted %s\n", oldVolumeName)
				}
			}
		} else {
			fmt.Println("    Old volumes retained. Delete manually with: docker volume rm <volume-name>")
		}
	}

	return migrated, skipped, errors
}

// findLegacyVolumeByKey returns the canonical legacy volume name for a given
// logical compose volume key. We only accept exact suffix matches to avoid
// migrating unrelated volumes.
func findLegacyVolumeByKey(volumes []string, volumeKey string) string {
	for _, vol := range volumes {
		if strings.HasSuffix(vol, "_"+volumeKey) {
			return vol
		}
	}
	return ""
}
