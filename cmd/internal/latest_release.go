package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func FetchLatestRelease() (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/repos/GhostManager/Ghostwriter/releases/latest", nil)
	if err != nil {
		return "", fmt.Errorf("Could not create request: %w", err)
	}
	req.Header.Add("User-Agent", "Ghostwriter-CLI")
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Could not send request: %w", err)
	}
	if res.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("Got status code %d", res.StatusCode))
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Could not read response body: %w", err)
	}

	var response githubReleaseResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("Could not parse response body: %w", err)
	}
	return response.Tag, nil
}

type githubReleaseResponse struct {
	Tag string `json:"tag_name"`
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
	if !docker.Env.env.GetBool("gwcli_auto_check_updates") {
		return
	}

	lastCheckFile := filepath.Join(docker.Dir, ".gwcli-last-update-check")
	lastCheckTime := readLastVersionCheck(lastCheckFile)
	now := time.Now().Unix()
	if lastCheckTime+(24*60*60) >= now {
		// Checked recently, do nothing
		return
	}

	err := os.WriteFile(lastCheckFile, []byte(strconv.FormatInt(now, 10)), 0666)
	if err != nil {
		fmt.Printf("[!] Could not write %s: %v", lastCheckFile, err)
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
