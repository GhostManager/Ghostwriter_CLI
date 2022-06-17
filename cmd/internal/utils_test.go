package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetCwdFromExe(t *testing.T) {
	cwd := GetCwdFromExe()
	if cwd == "" {
		t.Error("Could not get current working directory")
	}
}

func TestCheckPath(t *testing.T) {
	if !CheckPath("docker-compose") {
		t.Error("Did not find `docker-compose` in PATH")
	}
}

func TestRunBasicCmd(t *testing.T) {
	defer quietTests()()
	_, err := RunBasicCmd("docker-compose", []string{"--version"})
	if err != nil {
		t.Errorf("Could not run `docker-compose --version` with `RunBasicCmd()`: %s", err)
	}
}

func TestRunCmd(t *testing.T) {
	defer quietTests()()
	err := RunCmd("docker-compose", []string{"--version"})
	if err != nil {
		t.Errorf("Could not run `docker-compose --version` with `RunCmd()`: %s", err)
	}
}

func TestContains(t *testing.T) {
	if !Contains([]string{"a", "b", "c"}, "b") {
		t.Error("Expected `Contains()` to return true for `b`")
	}
	if Contains([]string{"a", "b", "c"}, "d") {
		t.Error("Expected `Contains()` to return false for `d`")
	}
}

func TestGetLocalGhostwriterVersion(t *testing.T) {

	// Mock the Ghostwriter VERSION file
	versionFile := filepath.Join(GetCwdFromExe(), "VERSION")
	f, err := os.Create(versionFile)
	if err != nil {
		t.Error(err)
	}

	defer f.Close()

	_, writeErr := f.WriteString("v3.0.0\n7 June 2022")
	if writeErr != nil {
		t.Error(writeErr)
	}

	// Test reading the version data from the file
	version, err := GetLocalGhostwriterVersion()
	if err != nil {
		t.Error(err)
	} else if version != "Ghostwriter v3.0.0 ( 7 June 2022 )\n" {
		t.Errorf("Expected `GetLocalGhostwriterVersion()` to return `Ghostwriter v3.0.0 ( 7 June 2022 )`, got `%s`", version)
	}
}

func TestGetRemoteGhostwriterVersion(t *testing.T) {
	// Test reading the version data from GitHub's API
	version, err := GetRemoteGhostwriterVersion()
	if err != nil {
		t.Errorf("Error getting remote Ghostwriter version: %s", err)
	}
	if !strings.Contains(version, "Ghostwriter v") {
		t.Errorf("Expected `GetRemoteGhostwriterVersion()` to return a string containing `Ghostwriter v...`, got `%s`", version)
	}
}
