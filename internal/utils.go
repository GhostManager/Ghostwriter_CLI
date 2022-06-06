package internal

// Various utilities used by other parts of the internal package
// Includes utilities for interacting with the file system

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Basic help info to display when no arguments are passed and for ``help`` case.
func DisplayHelp(version string, build string) {
	fmt.Printf("Ghostwriter-CLI ( v%s, %s ):\n", version, build)
	fmt.Println("********************************************************************")
	fmt.Println("*** source code: https://github.com/GhostManager/Ghostwriter_CLI ***")
	fmt.Println("********************************************************************")
	fmt.Println("  help")
	fmt.Println("    Displays this help information")
	fmt.Println("  install {dev|production}")
	fmt.Println("    Builds containers and performs first-time setup of Ghostwriter")
	fmt.Println("  build {dev|production}")
	fmt.Println("    Builds the containers for the given environment (only necessary for upgrades)")
	fmt.Println("  restart {dev|production}")
	fmt.Println("    Restarts all Ghostwriter services in the given environment")
	fmt.Println("  up {dev|production}")
	fmt.Println("    Bring up all Ghostwriter services in the given environment")
	fmt.Println("  down {dev|production}")
	fmt.Println("    Bring down all Ghostwriter services in the given environment")
	fmt.Println("  config")
	fmt.Println("    ** No parameters will dump the entire config **")
	fmt.Println("    get [varname ...]")
	fmt.Println("    set <var name> <var value>")
	fmt.Println("    allowhost <var hostname/address>")
	fmt.Println("    disallowhost <var hostname/address>")
	fmt.Println("  logs <container name>")
	fmt.Println("    Displays logs for the given container")
	fmt.Println("    Options: ghostwriter_{django|nginx|postgres|redis|graphql|queue}")
	fmt.Println("  running")
	fmt.Println("    Print a list of running Ghostwriter services")
	fmt.Println("  update")
	fmt.Println("    Displays version information for the local Ghostwriter installation and the latest stable release on GitHub")
	fmt.Println("  test")
	fmt.Println("    Runs Ghostwriter's unit tests in the development environment")
	fmt.Println("    Requires to `ghostwriter_cli install dev` to have been run first")
	fmt.Println("  version")
	fmt.Println("    Displays the version information at the top of this message")
}

// Get the current working directory based on ``ghostwriter-cli`` location.
func GetCwdFromExe() string {
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get path to current executable")
	}
	return filepath.Dir(exe)
}

// Determine if a given string is a valid filepath.
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

// Determine if a given string is a valid directory.
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

// Attempt to copy a ``src`` file to a ``dst`` destination.
// Reference: https://blog.depa.do/post/copy-files-and-directories-in-go
func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// Attempt to copy a ``src`` directory to a ``dst`` destination.
// Reference: https://blog.depa.do/post/copy-files-and-directories-in-go
func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// Ask user for confirmation and return a boolean.
// Reference: https://gist.github.com/r0l1/3dcbb0c8f6cfe9c66ab8008f55f8f28b
func AskConfirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/n]: ", prompt)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("[-] Failed to read user input\n")
			return false
		}
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" || input == "yes" {
			return true
		} else if input == "n" || input == "no" {
			return false
		}
	}
}

// Check the $PATH environment variable for a given ``cmd`` and return a ``bool``
// indicating if it exists.
func CheckPath(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// Execute a given command (``name``) with a list of arguments (``args``)
// and return a ``string`` with the output.
func RunBasicCmd(name string, args []string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	output := string(out[:])
	return output, err
}

// Execute a given command (``name``) with a list of arguments (``args``)
// and return stdout and stderr buffers.
func RunCmd(name string, args []string) error {
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

// Fetch the local Ghostwriter version from the ``VERSION`` file.
func GetLocalGhostwriterVersion() {
	if FileExists("./VERSION") {
		file, err := os.Open("VERSION")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Installed version: Ghostwriter %s ( %s )\n", lines[0], lines[1])
	} else {
		fmt.Println("[!] Could not read Ghostwriter's './VERSION' file")
	}
}

// Fetch the latest Ghostwriter version from GitHub's API.
func GetRemoteGhostwriterVersion() {
	baseUrl := "https://api.github.com/repos/GhostManager/Ghostwriter/releases/latest"
	client := http.Client{Timeout: time.Second * 2}
	resp, err := client.Get(baseUrl)
	if err != nil {
		log.Fatal(err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	var githubJson map[string]interface{}
	jsonErr := json.Unmarshal(body, &githubJson)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	publishedAt := githubJson["published_at"].(string)
	date, _ := time.Parse(time.RFC3339, publishedAt)
	fmt.Printf(
		"Latest stable version: Ghostwriter %s ( %02d %s %d )\n",
		githubJson["tag_name"], date.Day(), date.Month().String(), date.Year(),
	)
}

// Check if a slice of strings (``slice`` parameter) contains a given
// string (``search`` parameter).
func Contains(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}
