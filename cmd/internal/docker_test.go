package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateDockerComposeStatus(t *testing.T) {
	// Mock the Ghostwriter Docker YAML files
	localMockYaml := filepath.Join(GetCwdFromExe(), "local.yml")
	local, localErr := os.Create(localMockYaml)
	prodMockYaml := filepath.Join(GetCwdFromExe(), "production.yml")
	prod, prodErr := os.Create(prodMockYaml)
	assert.Equal(t, nil, localErr, "Expected `os.Create()` to return no error")
	assert.Equal(t, nil, prodErr, "Expected `os.Create()` to return no error")
	assert.True(t, FileExists(localMockYaml), "Expected `FileExists()` to return true")
	assert.True(t, FileExists(prodMockYaml), "Expected `FileExists()` to return true")

	defer local.Close()
	defer prod.Close()

	GetDockerInterface(ModeLocalDev)
}

// Note: The media backup and restore functions (RunDockerComposeMediaBackup and RunDockerComposeMediaRestore)
// require a full Docker environment with the appropriate volumes to test properly.
// These functions are tested through integration testing with the actual Ghostwriter deployment.

func TestVerifyVolumeExists(t *testing.T) {
	defer quietTests()()

	dockerInterface := GetDockerInterface(ModeLocalDev)

	// Test with a volume that definitely doesn't exist
	exists := dockerInterface.VerifyVolumeExists("nonexistent_test_volume_12345")
	assert.False(t, exists, "Expected nonexistent volume to return false")
}

func TestListVolumes(t *testing.T) {
	defer quietTests()()

	dockerInterface := GetDockerInterface(ModeLocalDev)

	// List all volumes with a filter that shouldn't match anything unusual
	volumes, err := dockerInterface.ListVolumes("test_filter_12345_nonexistent")
	assert.NoError(t, err, "Expected ListVolumes to return no error")
	assert.Equal(t, 0, len(volumes), "Expected no volumes with nonexistent filter")
}

func TestGetVolumeNameFromConfig(t *testing.T) {
	defer quietTests()()

	// Create a minimal mock compose file
	tempDir := t.TempDir()
	composeFile := filepath.Join(tempDir, "test-compose.yml")
	composeContent := `version: '3.8'
volumes:
  production_postgres_data:
    name: ghostwriter_production_postgres_data
  production_data:
    name: ghostwriter_production_data
`
	err := os.WriteFile(composeFile, []byte(composeContent), 0644)
	assert.NoError(t, err, "Expected to create test compose file")

	dockerInterface := GetDockerInterface(ModeLocalDev)
	dockerInterface.Dir = tempDir
	dockerInterface.ComposeFile = "test-compose.yml"

	// Test getting a volume name that exists in config
	volumeName, err := dockerInterface.GetVolumeNameFromConfig("production_postgres_data")
	assert.NoError(t, err, "Expected GetVolumeNameFromConfig to succeed")
	assert.Contains(t, volumeName, "postgres_data", "Expected volume name to contain postgres_data")
}
