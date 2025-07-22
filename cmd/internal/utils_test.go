package internal

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetCwdFromExe(t *testing.T) {
	cwd := GetCwdFromExe()
	assert.False(t, cwd == "", "Expected `GetCwdFromExe()` to return a non-empty string")
}

func TestCheckPath(t *testing.T) {
	assert.True(t, CheckPath("docker"), "Expected `CheckPath()` to find `docker` in `$PATH`")
}

func TestRunBasicCmd(t *testing.T) {
	defer quietTests()()
	_, err := RunBasicCmd(dockerCmd, []string{"--version"})
	assert.Equal(t, nil, err, "Expected `RunBasicCmd()` to return no error")
}

func TestRunCmd(t *testing.T) {
	defer quietTests()()
	err := RunCmd(dockerCmd, []string{"--version"})
	assert.Equal(t, nil, err, "Expected `RunCmd()` to return no error")
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
	version, _, err := GetRemoteVersion("GhostManager", "Ghostwriter")
	assert.NoError(t, err, "Expected `GetRemoteVersion()` to return no error")
	assert.True(
		t,
		strings.Contains(version, "Ghostwriter v"),
		"Expected `GetRemoteVersion()` to return a string containing `Ghostwriter v...`",
	)
}
