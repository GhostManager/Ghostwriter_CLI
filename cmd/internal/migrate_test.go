package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrateFile_Success(t *testing.T) {
	defer quietTests()()

	// Create temporary source and destination directories
	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Create a test file
	sourceFile := filepath.Join(sourceDir, "test.txt")
	content := []byte("test content")
	err := os.WriteFile(sourceFile, content, 0644)
	assert.NoError(t, err, "Expected to create source file successfully")

	// Migrate the file
	destFile := filepath.Join(destDir, "test.txt")
	migrated, err := MigrateFile(sourceFile, destFile, 0644, false)
	assert.NoError(t, err, "Expected MigrateFile to succeed")
	assert.True(t, migrated, "Expected file to be migrated")

	// Verify the file was copied
	assert.True(t, FileExists(destFile), "Expected destination file to exist")

	// Verify content matches
	destContent, err := os.ReadFile(destFile)
	assert.NoError(t, err, "Expected to read destination file")
	assert.Equal(t, content, destContent, "Expected content to match")

	// Verify permissions
	info, err := os.Stat(destFile)
	assert.NoError(t, err, "Expected to stat destination file")
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm(), "Expected correct permissions")
}

func TestMigrateFile_MissingSource(t *testing.T) {
	defer quietTests()()

	destDir := t.TempDir()
	sourceFile := filepath.Join(destDir, "nonexistent.txt")
	destFile := filepath.Join(destDir, "dest.txt")

	migrated, err := MigrateFile(sourceFile, destFile, 0644, false)
	assert.Error(t, err, "Expected error for missing source file")
	assert.False(t, migrated, "Expected file not to be migrated")
	assert.Contains(t, err.Error(), "does not exist", "Expected specific error message")
}

func TestMigrateFile_DestinationExists_SkipWithoutConfirm(t *testing.T) {
	defer quietTests()()

	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Create source and destination files
	sourceFile := filepath.Join(sourceDir, "test.txt")
	destFile := filepath.Join(destDir, "test.txt")

	err := os.WriteFile(sourceFile, []byte("source content"), 0644)
	assert.NoError(t, err, "Expected to create source file")

	err = os.WriteFile(destFile, []byte("existing content"), 0644)
	assert.NoError(t, err, "Expected to create destination file")

	// Migrate without confirmation
	migrated, err := MigrateFile(sourceFile, destFile, 0644, false)
	assert.NoError(t, err, "Expected no error when skipping")
	assert.False(t, migrated, "Expected file to be skipped")

	// Verify destination wasn't changed
	content, err := os.ReadFile(destFile)
	assert.NoError(t, err, "Expected to read destination file")
	assert.Equal(t, []byte("existing content"), content, "Expected destination unchanged")
}

func TestMigrateFile_PermissionHandling(t *testing.T) {
	defer quietTests()()

	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Test migrating with 0600 permissions
	sourceFile := filepath.Join(sourceDir, "secret.key")
	destFile := filepath.Join(destDir, "secret.key")

	err := os.WriteFile(sourceFile, []byte("secret"), 0644)
	assert.NoError(t, err, "Expected to create source file")

	migrated, err := MigrateFile(sourceFile, destFile, 0600, false)
	assert.NoError(t, err, "Expected MigrateFile to succeed")
	assert.True(t, migrated, "Expected file to be migrated")

	// Verify permissions
	info, err := os.Stat(destFile)
	assert.NoError(t, err, "Expected to stat destination file")
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm(), "Expected 0600 permissions")
}

func TestMigrateFile_CreatesDestinationDirectory(t *testing.T) {
	defer quietTests()()

	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Create source file
	sourceFile := filepath.Join(sourceDir, "test.txt")
	err := os.WriteFile(sourceFile, []byte("test"), 0644)
	assert.NoError(t, err, "Expected to create source file")

	// Migrate to nested destination that doesn't exist
	destFile := filepath.Join(destDir, "nested", "subdirs", "test.txt")
	migrated, err := MigrateFile(sourceFile, destFile, 0644, false)
	assert.NoError(t, err, "Expected MigrateFile to succeed")
	assert.True(t, migrated, "Expected file to be migrated")

	// Verify directory was created
	assert.True(t, DirExists(filepath.Join(destDir, "nested", "subdirs")), "Expected destination directories to be created")
	assert.True(t, FileExists(destFile), "Expected file to exist")
}

func TestMigrateDirectory_Success(t *testing.T) {
	defer quietTests()()

	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Create test files in source directory
	files := map[string]string{
		"file1.py":             "content 1",
		"file2.py":             "content 2",
		"subdir/file3.py":      "content 3",
		"subdir/deep/file4.py": "content 4",
	}

	for path, content := range files {
		fullPath := filepath.Join(sourceDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		assert.NoError(t, err, "Expected to create subdirectory")
		err = os.WriteFile(fullPath, []byte(content), 0644)
		assert.NoError(t, err, "Expected to create file %s", path)
	}

	// Migrate directory
	result, err := MigrateDirectory(sourceDir, destDir, false)
	assert.NoError(t, err, "Expected MigrateDirectory to succeed")
	assert.Equal(t, 4, result.Migrated, "Expected 4 files to be migrated")
	assert.Equal(t, 0, result.Skipped, "Expected 0 files to be skipped")
	assert.Equal(t, 0, result.Failed, "Expected 0 files to fail")

	// Verify all files were migrated
	for path, content := range files {
		destPath := filepath.Join(destDir, path)
		assert.True(t, FileExists(destPath), "Expected %s to exist", path)
		destContent, err := os.ReadFile(destPath)
		assert.NoError(t, err, "Expected to read %s", path)
		assert.Equal(t, content, string(destContent), "Expected content to match for %s", path)
	}
}

func TestMigrateDirectory_MissingSource(t *testing.T) {
	defer quietTests()()

	sourceDir := filepath.Join(t.TempDir(), "nonexistent")
	destDir := t.TempDir()

	result, err := MigrateDirectory(sourceDir, destDir, false)
	assert.Error(t, err, "Expected error for missing source directory")
	assert.NotNil(t, result, "Expected result even on error")
	assert.Contains(t, err.Error(), "does not exist", "Expected specific error message")
}

func TestMigrateDirectory_EmptySource(t *testing.T) {
	defer quietTests()()

	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Migrate empty directory
	result, err := MigrateDirectory(sourceDir, destDir, false)
	assert.NoError(t, err, "Expected MigrateDirectory to succeed for empty directory")
	assert.Equal(t, 0, result.Migrated, "Expected 0 files to be migrated")
	assert.Equal(t, 0, result.Skipped, "Expected 0 files to be skipped")
	assert.Equal(t, 0, result.Failed, "Expected 0 files to fail")
}

func TestMigrateDirectory_SkipsExistingFiles(t *testing.T) {
	defer quietTests()()

	sourceDir := t.TempDir()
	destDir := t.TempDir()

	// Create files in source
	err := os.WriteFile(filepath.Join(sourceDir, "file1.py"), []byte("source 1"), 0644)
	assert.NoError(t, err, "Expected to create source file 1")
	err = os.WriteFile(filepath.Join(sourceDir, "file2.py"), []byte("source 2"), 0644)
	assert.NoError(t, err, "Expected to create source file 2")

	// Create one file in destination
	err = os.WriteFile(filepath.Join(destDir, "file1.py"), []byte("existing 1"), 0644)
	assert.NoError(t, err, "Expected to create destination file")

	// Migrate without confirmation
	result, err := MigrateDirectory(sourceDir, destDir, false)
	assert.NoError(t, err, "Expected MigrateDirectory to succeed")
	assert.Equal(t, 1, result.Migrated, "Expected 1 file to be migrated")
	assert.Equal(t, 1, result.Skipped, "Expected 1 file to be skipped")
	assert.Equal(t, 0, result.Failed, "Expected 0 files to fail")

	// Verify existing file wasn't changed
	content, err := os.ReadFile(filepath.Join(destDir, "file1.py"))
	assert.NoError(t, err, "Expected to read existing file")
	assert.Equal(t, "existing 1", string(content), "Expected existing file unchanged")

	// Verify new file was migrated
	content, err = os.ReadFile(filepath.Join(destDir, "file2.py"))
	assert.NoError(t, err, "Expected to read new file")
	assert.Equal(t, "source 2", string(content), "Expected new file migrated")
}

func TestMigrationResult_AddError(t *testing.T) {
	result := &MigrationResult{}
	assert.Equal(t, 0, result.Failed, "Expected initial failed count to be 0")
	assert.Equal(t, 0, len(result.Errors), "Expected initial errors to be empty")

	result.AddError(assert.AnError)
	assert.Equal(t, 1, result.Failed, "Expected failed count to increment")
	assert.Equal(t, 1, len(result.Errors), "Expected one error in list")
	assert.Equal(t, assert.AnError, result.Errors[0], "Expected error to be stored")
}
