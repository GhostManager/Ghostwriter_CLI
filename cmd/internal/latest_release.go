package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// FetchLatestRelease fetches the latest Ghostwriter release tag from GitHub.
// This is a convenience wrapper around GetRemoteVersion for the specific case
// of checking the Ghostwriter repository.
func FetchLatestRelease() (string, error) {
	tag, _, err := GetRemoteVersion("GhostManager", "Ghostwriter")
	return tag, err
}

func readLastVersionCheck(file string) int64 {
	lastDateBytes, err := os.ReadFile(file)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Printf("[!] Could not read %s file: %v\n", file, err)
		}
		return 0
	}
	lastDateText := string(lastDateBytes)
	lastDate, err := strconv.ParseInt(lastDateText, 10, 64)
	if err != nil {
		fmt.Printf("[!] Could not read %s file: %v\n", file, err)
		return 0
	}
	return lastDate
}

func CheckLatestVersionNag(docker *DockerInterface) {
	if !docker.Env.GetBool("gwcli_auto_check_updates") {
		return
	}

	lastCheckFile := filepath.Join(docker.Dir, ".gwcli-last-update-check")
	lastCheckTime := readLastVersionCheck(lastCheckFile)
	now := time.Now().Unix()
	if lastCheckTime+(24*60*60) >= now {
		// Checked recently, do nothing
		return
	}

	err := os.WriteFile(lastCheckFile, []byte(strconv.FormatInt(now, 10)), 0600)
	if err != nil {
		fmt.Printf("[!] Could not write %s: %v\n", lastCheckFile, err)
	}

	localVersion, err := docker.GetVersion()
	if err != nil {
		fmt.Printf("[!] Could not get local version: %v\n", err)
		return
	}
	remoteVersion, err := FetchLatestRelease()
	if err != nil {
		fmt.Printf("[!] Could not get latest released version: %v\n", err)
		return
	}

	if localVersion != remoteVersion {
		fmt.Printf("[!] The latest release of Ghostwriter is version %s - the currently installed version is %s\n", remoteVersion, localVersion)
		if docker.ManageComposeFile {
			fmt.Print("[!] Run the `install` command to update to the latest version\n")
		}
	}
}
