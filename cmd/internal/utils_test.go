package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCwdFromExe(t *testing.T) {
	cwd := GetCwdFromExe()
	assert.False(t, cwd == "", "Expected `GetCwdFromExe()` to return a non-empty string")
}

func TestCheckPath(t *testing.T) {
	dockerFound := CheckPath("docker")
	if !dockerFound {
		dockerFound = CheckPath("podman")
	}
	assert.True(t, dockerFound, "Expected `CheckPath()` to find `docker` or `podman` in `$PATH`")
}

func TestContains(t *testing.T) {
	assert.True(t, Contains([]string{"a", "b", "c"}, "b"), "Expected `Contains()` to return true")
	assert.False(t, Contains([]string{"a", "b", "c"}, "d"), "Expected `Contains()` to return false")
}

func TestGetLocalGhostwriterVersion(t *testing.T) {
	// Mock the Ghostwriter VERSION file
	versionFile := filepath.Join(GetCwdFromExe(), "VERSION")
	f, err := os.Create(versionFile)
	assert.NoError(t, err, "Expected `os.Create()` to return no error")

	defer f.Close()

	_, writeErr := f.WriteString("v3.0.0\n22 June 2022")
	assert.NoError(t, writeErr, "Expected `f.WriteString()` to return no error")

	// Test reading the version data from the file
	version, err := GetLocalGhostwriterVersion()
	assert.NoError(t, err, "Expected `GetLocalGhostwriterVersion()` to return no error")
	assert.Equal(
		t,
		"Ghostwriter v3.0.0 (22 June 2022)",
		version,
		"Expected `GetLocalGhostwriterVersion()` to return `Ghostwriter v3.0.0 (22 June 2022)`",
	)
}

func TestGetRemoteVersion(t *testing.T) {
	// Test reading the version data from GitHub's API
	version, url, err := GetRemoteVersion("GhostManager", "Ghostwriter")
	
	// Skip test if we hit GitHub API rate limiting (common in CI environments)
	if err != nil && strings.Contains(err.Error(), "403") {
		t.Skip("Skipping test due to GitHub API rate limiting (HTTP 403)")
	}
	
	assert.NoError(t, err, "Expected `GetRemoteVersion()` to return no error")
	assert.NotEmpty(t, version, "Expected `GetRemoteVersion()` to return a non-empty version string")
	assert.True(
		t,
		strings.HasPrefix(version, "v"),
		"Expected `GetRemoteVersion()` to return a version string starting with `v`",
	)
	assert.NotEmpty(t, url, "Expected `GetRemoteVersion()` to return a non-empty URL")
	assert.True(
		t,
		strings.Contains(url, "github.com/GhostManager/Ghostwriter"),
		"Expected `GetRemoteVersion()` to return a URL containing the Ghostwriter repository",
	)
}
