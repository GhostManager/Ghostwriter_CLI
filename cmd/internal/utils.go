package internal

// Various utilities used by other parts of the internal package
// Includes utilities for interacting with the file system

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GetCwdFromExe gets the current working directory based on "ghostwriter-cli" location.
func GetCwdFromExe() string {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get path to current executable")
	}
	return filepath.Dir(exe)
}

// FileExists determines if a given string is a valid filepath.
// Reference: https://golangcode.com/check-if-a-file-exists/
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return !info.IsDir()
}

// DirExists determines if a given string is a valid directory.
// Reference: https://golangcode.com/check-if-a-file-exists/
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return info.IsDir()
}

// CheckPath checks the $PATH environment variable for a given "cmd" and return a "bool"
// indicating if it exists.
func CheckPath(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// GetLocalGhostwriterVersion fetches the local Ghostwriter version from the "VERSION" file.
func GetLocalGhostwriterVersion() (string, error) {
	var output string

	versionFile := filepath.Join(GetCwdFromExe(), "VERSION")
	if FileExists(versionFile) {
		file, err := os.Open(versionFile)
		if err != nil {
			return output, err
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return output, err
		}

		output = fmt.Sprintf("Ghostwriter %s (%s)", lines[0], lines[1])
	} else {
		output = "Could not read Ghostwriter's `VERSION` file"
	}

	return output, nil
}

// githubReleaseResponse represents the GitHub API response for a release.
type githubReleaseResponse struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
}

// GetRemoteVersion fetches the latest version information from GitHub's API for the given repository.
// It includes proper headers to avoid rate limiting and uses struct-based unmarshaling for type safety.
func GetRemoteVersion(owner string, repository string) (string, string, error) {
	baseUrl := "https://api.github.com/repos/" + owner + "/" + repository + "/releases/latest"
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		return "", "", fmt.Errorf("could not create request: %w", err)
	}

	// Add GitHub API headers to avoid rate limiting and ensure proper API version
	req.Header.Add("User-Agent", "Ghostwriter-CLI")
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("could not send request: %w", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("could not read response body: %w", err)
	}

	var response githubReleaseResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", "", fmt.Errorf("could not parse response body: %w", err)
	}

	if response.TagName == "" {
		return "", "", fmt.Errorf("tag_name field missing or empty in GitHub API response")
	}
	if response.HtmlUrl == "" {
		return "", "", fmt.Errorf("html_url field missing or empty in GitHub API response")
	}

	return response.TagName, response.HtmlUrl, nil
}

// Contains checks if a slice of strings ("slice" parameter) contains a given
// string ("search" parameter).
func Contains(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// Silence any output from tests.
// Place `defer quietTests()()` after test declarations.
// Ref: https://stackoverflow.com/a/58720235
func quietTests() func() {
	null, _ := os.Open(os.DevNull)
	sout := os.Stdout
	serr := os.Stderr
	os.Stdout = null
	os.Stderr = null
	log.SetOutput(null)
	return func() {
		defer null.Close()
		os.Stdout = sout
		os.Stderr = serr
		log.SetOutput(os.Stderr)
	}
}

// AskForConfirmation asks the user for confirmation. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user.
// Original source: https://gist.github.com/r0l1/3dcbb0c8f6cfe9c66ab8008f55f8f28b
func AskForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

// MigrationResult tracks the outcome of a migration operation.
type MigrationResult struct {
	Migrated int
	Skipped  int
	Failed   int
	Errors   []error
}

// AddError adds an error to the migration result and increments the failure counter.
func (r *MigrationResult) AddError(err error) {
	r.Failed++
	r.Errors = append(r.Errors, err)
}

// MigrateFile copies a file from source to destination with atomic write and permission handling.
// If confirm is true and the destination exists, it asks the user for confirmation to overwrite.
// Returns true if the file was migrated, false if it was skipped.
func MigrateFile(sourcePath, destPath string, perm os.FileMode, confirm bool) (bool, error) {
	// Check if source exists
	if !FileExists(sourcePath) {
		return false, fmt.Errorf("source file does not exist: %s", sourcePath)
	}

	// Check if destination exists
	if FileExists(destPath) {
		if confirm {
			prompt := fmt.Sprintf("File %s already exists. Overwrite?", filepath.Base(destPath))
			if !AskForConfirmation(prompt) {
				return false, nil // Skipped by user
			}
		} else {
			return false, nil // Skip without confirmation
		}
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return false, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read source file
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return false, fmt.Errorf("failed to read source file: %w", err)
	}

	// Write to temporary file first (atomic write pattern)
	tempFile, err := os.CreateTemp(destDir, ".migrate-*")
	if err != nil {
		return false, fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath) // Clean up temp file if something goes wrong

	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		return false, fmt.Errorf("failed to write temporary file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return false, fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Set permissions
	if err := os.Chmod(tempPath, perm); err != nil {
		return false, fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, destPath); err != nil {
		return false, fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return true, nil
}

// MigrateDirectory recursively copies all files from sourceDir to destDir.
// If confirm is true, it asks for confirmation before overwriting existing files.
// Returns a MigrationResult with statistics about the operation.
func MigrateDirectory(sourceDir, destDir string, confirm bool) (*MigrationResult, error) {
	result := &MigrationResult{}

	// Check if source directory exists
	if !DirExists(sourceDir) {
		return result, fmt.Errorf("source directory does not exist: %s", sourceDir)
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return result, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Walk the source directory
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.AddError(fmt.Errorf("failed to access path %s: %w", path, err))
			return nil // Continue walking
		}

		// Skip directories (we only migrate files)
		if info.IsDir() {
			return nil
		}

		// Skip .gitignore files
		if info.Name() == ".gitignore" {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			result.AddError(fmt.Errorf("failed to calculate relative path for %s: %w", path, err))
			return nil // Continue walking
		}

		// Calculate destination path
		destPath := filepath.Join(destDir, relPath)

		// Use standard file permissions for migrated files
		perm := os.FileMode(0644)

		// Migrate the file
		migrated, err := MigrateFile(path, destPath, perm, confirm)
		if err != nil {
			result.AddError(fmt.Errorf("failed to migrate %s: %w", relPath, err))
			return nil // Continue walking
		}

		if migrated {
			result.Migrated++
		} else {
			result.Skipped++
		}

		return nil
	})

	if err != nil {
		return result, fmt.Errorf("failed to walk source directory: %w", err)
	}

	return result, nil
}
