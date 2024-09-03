package internal

// Various utilities used by other parts of the internal package
// Includes utilities for interacting with the file system

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// HealthIssue is a custom type for storing healthcheck output.
type HealthIssue struct {
	Type    string
	Service string
	Message string
}

type HealthIssues []HealthIssue

func (c HealthIssues) Len() int {
	return len(c)
}

func (c HealthIssues) Less(i, j int) bool {
	return c[i].Service < c[j].Service
}

func (c HealthIssues) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// GetCwdFromExe gets the current working directory based on “ghostwriter-cli“ location.
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

// CheckPath checks the $PATH environment variable for a given “cmd“ and return a “bool“
// indicating if it exists.
func CheckPath(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// RunBasicCmd executes a given command (“name“) with a list of arguments (“args“)
// and return a “string“ with the output.
func RunBasicCmd(name string, args []string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	output := string(out[:])
	return output, err
}

// RunCmd executes a given command (“name“) with a list of arguments (“args“)
// and return stdout and stderr buffers.
func RunCmd(name string, args []string) error {
	// If the command is ``docker``, prepend ``compose`` to the args
	if name == "docker" {
		args = append([]string{"compose"}, args...)
	}
	path, err := exec.LookPath(name)
	if err != nil {
		log.Fatalf("`%s` is not installed or not available in the current PATH variable", name)
	}
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get path to current executable")
	}
	exePath := filepath.Dir(exe)
	command := exec.Command(path, args...)
	command.Dir = exePath

	stdout, err := command.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe for running `%s`", name)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to get stderr pipe for running `%s`", name)
	}

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stdoutScanner.Scan() {
			fmt.Printf("%s\n", stdoutScanner.Text())
		}
	}()
	go func() {
		for stderrScanner.Scan() {
			fmt.Printf("%s\n", stderrScanner.Text())
		}
	}()
	err = command.Start()
	if err != nil {
		log.Fatalf("Error trying to start `%s`: %v\n", name, err)
	}
	err = command.Wait()
	if err != nil {
		fmt.Printf("[-] Error from `%s`: %v\n", name, err)
		return err
	}
	return nil
}

// GetLocalGhostwriterVersion fetches the local Ghostwriter version from the “VERSION“ file.
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

		output = fmt.Sprintf("Ghostwriter %s ( %s )\n", lines[0], lines[1])
	} else {
		output = "Could not read Ghostwriter's `VERSION` file\n"
	}

	return output, nil
}

// GetRemoteGhostwriterVersion fetches the latest Ghostwriter version from GitHub's API.
func GetRemoteGhostwriterVersion() (string, error) {
	var output string

	baseUrl := "https://api.github.com/repos/GhostManager/Ghostwriter/releases/latest"
	client := http.Client{Timeout: time.Second * 2}
	resp, err := client.Get(baseUrl)
	if err != nil {
		return "", err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", readErr
	}

	var githubJson map[string]interface{}
	jsonErr := json.Unmarshal(body, &githubJson)
	if jsonErr != nil {
		return "", jsonErr
	}

	publishedAt := githubJson["published_at"].(string)
	date, _ := time.Parse(time.RFC3339, publishedAt)
	output = fmt.Sprintf(
		"Ghostwriter %s ( %02d %s %d )\n",
		githubJson["tag_name"], date.Day(), date.Month().String(), date.Year(),
	)

	return output, nil
}

// Contains checks if a slice of strings (“slice“ parameter) contains a given
// string (“search“ parameter).
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

// CheckGhostwriterHealth fetches the latest health reports from Ghostwriter's status API endpoint.
func CheckGhostwriterHealth(dev bool) (HealthIssues, error) {
	var issues HealthIssues

	protocol := "https"
	port := "443"
	if dev {
		protocol = "http"
		port = "8000"
	}

	baseUrl := protocol + "://localhost:" + port + "/status/"
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := http.Client{Timeout: time.Second * 2, Transport: transport}

	req, err := http.NewRequest(http.MethodGet, baseUrl, nil)
	if err != nil {
		return issues, err
	}

	req.Header.Set("Accept", "application/json")

	res, getErr := client.Do(req)

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		return issues, errors.New("Non-OK HTTP status suggests an issue with the Django or Nginx services (Code " + strconv.Itoa(res.StatusCode) + ")")
	}
	if getErr != nil {
		return issues, getErr
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return issues, readErr
	}

	var results map[string]interface{}
	jsonErr := json.Unmarshal(body, &results)
	if jsonErr != nil {
		return issues, jsonErr
	}

	for key := range results {
		if results[key] != "working" {
			issues = append(issues, HealthIssue{"Service", key, results[key].(string)})
		}
	}

	return issues, nil
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
