package internal

import (
	"context"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/goccy/go-yaml"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// Vars for tracking the list of Ghostwriter images
// Used for filtering the list of containers returned by the Docker client
var (
	ProdImages = []string{
		"ghostwriter_production_django", "ghostwriter_production_nginx",
		"ghostwriter_production_redis", "ghostwriter_production_postgres",
		"ghostwriter_production_graphql", "ghostwriter_production_queue",
		"ghostwriter_production_collab_server",
	}
	DevImages = []string{
		"ghostwriter_local_django", "ghostwriter_local_redis",
		"ghostwriter_local_postgres", "ghostwriter_local_graphql",
		"ghostwriter_local_queue", "ghostwriter_local_collab_server",
		"ghostwriter_local_frontend",
	}
)

// Run mode - specifies where to get dockerfiles and whether to run dev or prod
type DockerMode string

const (
	// Use source in exe's directory in dev mode
	ModeLocalDev DockerMode = "local-dev"
	// Use source in exe's directory in prod mode
	ModeLocalProd DockerMode = "local-prod"
	// Download and manage dockerfiles and run in prod mode
	ModeProd DockerMode = "prod"
)

var AllModes = []string{string(ModeLocalDev), string(ModeLocalProd), string(ModeProd)}

// cobra pvalue.Value implementation for argument parsing
func (e *DockerMode) String() string {
	return string(*e)
}
func (e *DockerMode) Set(v string) error {
	if !slices.Contains(AllModes, v) {
		return errors.New("must be one of: " + strings.Join(AllModes, ", "))
	}
	*e = DockerMode(v)
	return nil
}
func (e *DockerMode) Type() string {
	return "DockerMode"
}

type DockerInterface struct {
	// Directory that docker compose file resides in
	Dir string
	// Docker compose filename to use, without directory
	ComposeFile string
	// Use development image names and environment settings instead of production ones
	UseDevInfra bool
	// Whether GW-CLI should download and write the compose file
	ManageComposeFile bool
	// Command to use, either docker or podman
	command string
	// Daemon client, lazily initialized
	client *client.Client
	// Docker environmental variables
	Env *GWEnvironment
	// Compose project name, lazily fetched
	composeProjectName string
}

// Gets the directory that the docker-compose and other files are in, depending on the run mode
func GetDockerDirFromMode(mode DockerMode) string {
	if mode == ModeProd {
		dir, err := xdg.DataFile("ghostwriter/prod.yml")
		if err != nil {
			log.Fatalf("Could not get data directory: %s\n", err)
		}
		dir = filepath.Dir(dir)
		if err := os.MkdirAll(dir, 0700); err != nil {
			log.Fatalf("Could not create directory %s: %s\n", dir, err)
		}
		return dir
	}
	return GetCwdFromExe()
}

// Gets the docker interface, checking how to run docker/podman, etc
func GetDockerInterface(mode DockerMode) *DockerInterface {
	fmt.Println("[+] Checking the status of Docker and the Compose plugin...")
	// Check for ``docker`` first because it's required for everything to come
	dockerExists := CheckPath("docker")
	dockerCmd := "docker"
	if !dockerExists {
		podmanExists := CheckPath("podman")
		if podmanExists {
			fmt.Println("[+] Docker is not installed, but Podman is installed. Using Podman as a Docker alternative.")
			dockerCmd = "podman"
		} else {
			log.Fatalln("Neither Docker nor Podman is installed on this system, so please install Docker or Podman (in Docker compatibility mode) and try again.")
		}
	}

	// Check if the Docker Engine is running
	_, engineErr := exec.Command(dockerCmd, "info").Output()
	if engineErr != nil {
		if strings.Contains(strings.ToLower(engineErr.Error()), "permission denied") {
			log.Fatalf("%s is installed, but you don't have permission to talk to the daemon (Try running with sudo or adjusting your group membership)", dockerCmd)
		} else {
			log.Fatalf("%s is installed on this system, but the daemon may not be running", dockerCmd)
		}
	}

	// Check for the ``compose`` plugin as our first choice
	_, composeErr := exec.Command(dockerCmd, "compose", "version").Output()
	if composeErr != nil {
		// Check if the deprecated v1 script is installed
		composeScriptExists := CheckPath("docker-compose")
		if composeScriptExists {
			fmt.Println("[!] The deprecated `docker-compose` v1 script was detected on your system")
			fmt.Println("[!] Docker has deprecated v1 and this CLI tool no longer supports it")
			log.Fatalln("Please upgrade to Docker Compose v2 and try again: https://docs.docker.com/compose/install/")
		} else {
			log.Fatalln("Docker Compose is not installed, so please install it and try again: https://docs.docker.com/compose/install/")
		}
	}

	dir := GetDockerDirFromMode(mode)

	var file string
	switch mode {
	case ModeLocalDev:
		file = "local.yml"
	case ModeLocalProd:
		file = "production.yml"
	case ModeProd:
		file = "docker-compose.yml"
	default:
		panic("Unrecognized mode - this is a bug")
	}

	// Bail out if a compose file isn't available.
	// Otherwise, we'll get a confusing error message from the `compose` plugin
	if !FileExists(filepath.Join(dir, file)) {
		if mode == ModeProd {
			log.Fatalf("Ghostwriter is not installed - please run the `install` command first.")
		} else {
			log.Fatalf("Ghostwriter CLI must be run in the same directory as the %s file", file)
		}
	}

	env, err := ReadEnv(dir)
	if err != nil {
		log.Fatalf("Could not load environment file: %s\n", err)
	}

	if mode == ModeLocalDev {
		env.SetDev()
	} else {
		env.SetProd()
	}

	return &DockerInterface{
		Dir:                dir,
		ComposeFile:        file,
		UseDevInfra:        mode == ModeLocalDev,
		ManageComposeFile:  mode == ModeProd,
		command:            dockerCmd,
		client:             nil,
		Env:                env,
		composeProjectName: "",
	}
}

// Runs docker/podman with the specified additional arguments, in the proper CWD with the env and compose files.
// Basis for most of the other Run commands.
func (this *DockerInterface) RunCmd(args ...string) error {
	path, err := exec.LookPath(this.command)
	if err != nil {
		log.Fatalf("`%s` is not installed or not available in the current PATH variable", this.command)
	}
	command := exec.Command(path, args...)
	command.Dir = this.Dir
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err = command.Start()
	if err != nil {
		log.Fatalf("Error trying to start `%s`: %v\n", this.command, err)
	}
	err = command.Wait()
	if err != nil {
		fmt.Printf("[-] Error from `%s`: %v\n", this.command, err)
		return err
	}
	return nil
}

// Similar to `RunCmd` but returns stdout
func (this *DockerInterface) RunCmdWithOutput(args ...string) (string, error) {
	path, err := exec.LookPath(this.command)
	if err != nil {
		log.Fatalf("`%s` is not installed or not available in the current PATH variable", this.command)
	}
	command := exec.Command(path, args...)
	command.Dir = this.Dir
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	out, err := command.Output()
	output := string(out[:])
	return output, err
}

// Runs a `docker compose` subcommand, pointing to the configured compose file, with additional arguments.
func (this *DockerInterface) RunComposeCmd(args ...string) error {
	args = append([]string{"compose", "-f", this.ComposeFile}, args...)
	return this.RunCmd(args...)
}

// Bring all containers up
func (this *DockerInterface) Up() error {
	fmt.Printf("[+] Running `%s` to bring up the containers with %s...\n", this.command, this.ComposeFile)
	return this.RunComposeCmd("up", "-d")
}

// Options for `Down`
type DownOptions struct {
	// Pass `--volumes` to delete the project's volumes as well (will lose data!)
	Volumes bool
	// Pass `--remove-orphans` to delete orphaned service containers
	RemoveOrphans bool
}

// Take down all containers. `opts` are optional
func (this *DockerInterface) Down(opts *DownOptions) error {
	fmt.Printf("[+] Running `%s` to take down the containers with %s...\n", this.command, this.ComposeFile)
	args := []string{"down"}
	if opts != nil {
		if opts.Volumes {
			args = append(args, "--volumes")
		}
		if opts.RemoveOrphans {
			args = append(args, "--remove-orphans")
		}
	}
	return this.RunComposeCmd(args...)
}

// Gets the docker compose project name
func (this *DockerInterface) GetComposeProjectName() string {
	if this.composeProjectName != "" {
		return this.composeProjectName
	}

	out, err := this.RunCmdWithOutput("compose", "-f", this.ComposeFile, "config", "--format", "json")
	if err != nil {
		log.Fatalf("Could not get docker compose project info: %s\n", err)
	}

	path, err := yaml.PathString("$.name")
	if err != nil {
		log.Fatalf("Could not parse yaml path. This is a bug. %s\n", err)
	}

	var name string
	err = path.Read(strings.NewReader(out), &name)
	if err != nil {
		log.Fatalf("Could not get docker compose project name: %s\n", err)
	}

	this.composeProjectName = name
	return name
}

// Container is a custom type for storing container information similar to output from "docker containers ls".
type Container struct {
	ID     string
	Image  string
	Status string
	Ports  []container.PortSummary
	Name   string
}

// Containers is a collection of Container structs
type Containers []Container

// Len returns the length of a Containers struct
func (c Containers) Len() int {
	return len(c)
}

// Less determines if one Container is less than another Container
func (c Containers) Less(i, j int) bool {
	return c[i].Image < c[j].Image
}

// Swap exchanges the position of two Container values in a Containers struct
func (c Containers) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Gets a list of all running docker containers (including outside of this project)
func (this *DockerInterface) GetRunning() Containers {
	var running Containers

	cli, err := this.GetDaemonClient()
	if err != nil {
		log.Fatalf("Failed to get client connection to Docker: %v", err)
	}
	containers, err := cli.ContainerList(context.Background(), client.ContainerListOptions{
		All: false,
	})
	if err != nil {
		log.Fatalf("Failed to get container list from Docker: %v", err)
	}

	for _, container := range containers.Items {
		if Contains(DevImages, container.Image) || Contains(ProdImages, container.Image) {
			running = append(running, Container{
				container.ID, container.Image, container.Status, container.Ports, container.Labels["name"],
			})
		}
	}

	return running
}

// Gets logs from a container
func (this *DockerInterface) FetchLogs(containerName string, lines string) []string {
	var logs []string
	cli, err := this.GetDaemonClient()
	if err != nil {
		log.Fatalf("Failed to get client in logs: %v", err)
	}
	containers, err := cli.ContainerList(context.Background(), client.ContainerListOptions{})
	if err != nil {
		log.Fatalf("Failed to get container list: %v", err)
	}
	if len(containers.Items) > 0 {
		for _, container := range containers.Items {
			if container.Labels["name"] == containerName || containerName == "all" || container.Labels["name"] == "ghostwriter_"+containerName {
				logs = append(logs, fmt.Sprintf("\n*** Logs for `%s` ***\n\n", container.Labels["name"]))
				reader, err := cli.ContainerLogs(context.Background(), container.ID, client.ContainerLogsOptions{
					ShowStdout: true,
					ShowStderr: true,
					Tail:       lines,
				})
				if err != nil {
					log.Fatalf("Failed to get container logs: %v", err)
				}
				// Reference: https://medium.com/@dhanushgopinath/reading-docker-container-logs-with-golang-docker-engine-api-702233fac044
				p := make([]byte, 8)
				_, err = reader.Read(p)
				for err == nil {
					content := make([]byte, binary.BigEndian.Uint32(p[4:]))
					reader.Read(content)
					logs = append(logs, string(content))
					_, err = reader.Read(p)
				}
				reader.Close()
			}
		}

		if len(logs) == 0 {
			logs = append(logs, fmt.Sprintf("\n*** No logs found for requested container '%s' ***\n", containerName))
		}
	} else {
		fmt.Println("Failed to find that container running (try checking with `./ghostwriter-cli running`)")
	}
	return logs
}

// Determine if the container with the specified name is running
func (this *DockerInterface) IsServiceRunning(containerName string) bool {
	projectName := this.GetComposeProjectName()
	name := fmt.Sprintf("%s-%s-1", projectName, containerName)

	out, err := this.RunCmdWithOutput("inspect", "-f", "json", name)
	if err != nil {
		log.Fatalf("Could not get status of container %s: %s\n", name, err)
	}

	path, err := yaml.PathString("$[0].State.Running")
	if err != nil {
		log.Fatalf("Could not parse yaml path. This is a bug. %s\n", err)
	}

	var running bool
	err = path.Read(strings.NewReader(out), &running)
	if err != nil {
		log.Fatalf("Could not get status of %s: %s\n", name, err)
	}

	return running
}

// Determine if the Django application has completed startup based on
// the "Application startup complete" log message.
func (this *DockerInterface) IsDjangoStarted() bool {
	expectedString := "Application startup complete"
	logs := this.FetchLogs("ghostwriter_django", "500")
	for _, entry := range logs {
		result := strings.Contains(entry, expectedString)
		if result {
			return true
		}
	}
	return false
}

// Check if PostgreSQL is having trouble starting due to a password mismatch.
func (this *DockerInterface) IsPostgresStarted() bool {
	expectedString := "Password does not match for user"
	logs := this.FetchLogs("ghostwriter_postgres", "100")
	for _, entry := range logs {
		result := strings.Contains(entry, expectedString)
		if result {
			return true
		}
	}
	return false
}

// Determine if the Ghostwriter application has completed startup
func (this *DockerInterface) WaitForDjango() bool {
	// Wait for ghostwriter to start running
	fmt.Println("[+] Waiting for Django application startup to complete...")
	counter := 0
	for {
		if !this.IsServiceRunning("django") {
			fmt.Print("\n")
			log.Fatalf("Django container exited unexpectedly. Check the logs in docker for the django container")
		}
		if this.IsDjangoStarted() {
			fmt.Print("\n[+] Django application started\n")
			return true
		}
		if this.IsPostgresStarted() {
			fmt.Print("\n")
			log.Fatalf("PostgreSQL cannot start because of a password mismatch. Please read: https://www.ghostwriter.wiki/getting-help/faq#ghostwriter-cli-reports-an-issue-with-postgresql")
		}

		if counter > 120 {
			fmt.Print("\n")
			log.Fatalf("Django did not start after 120 seconds.")
		}

		fmt.Print(".")
		time.Sleep(1 * time.Second)
		counter++
	}
}

// Runs the django manage.py script, with the specified arguments
func (this *DockerInterface) RunDjangoManageCommand(args ...string) error {
	args = append([]string{"run", "--rm", "django", "python", "manage.py"}, args...)
	return this.RunComposeCmd(args...)
}

// Connects to the docker daemon
func (this *DockerInterface) GetDaemonClient() (*client.Client, error) {
	if this.client != nil {
		return this.client, nil
	}

	client, err := client.New(client.FromEnv, client.WithAPIVersionNegotiation())
	this.client = client
	return this.client, err
}

// Gets the currently installed version of Ghostwriter
func (this *DockerInterface) GetVersion() (string, error) {
	if this.ManageComposeFile {
		// get the version embedded in the compose file
		out, err := this.RunCmdWithOutput("compose", "-f", this.ComposeFile, "config", "--images")
		if err != nil {
			return "", fmt.Errorf("Could not list docker images: %w", err)
		}
		re := regexp.MustCompile(`^[^\:]+:([^\n]+)`)
		captures := re.FindStringSubmatch(out)
		if captures == nil || len(captures) < 2 {
			return "", fmt.Errorf("Could not find version number in docker images")
		}
		return captures[1], nil
	}

	// get the version in the source tree's VERSION file
	versionFileBytes, err := os.ReadFile(filepath.Join(this.Dir, "VERSION"))
	if err != nil {
		return "", fmt.Errorf("Could not read VERSION file: %w", err)
	}
	versionFile := string(versionFileBytes)
	return strings.Split(versionFile, "\n")[0], nil
}

// GetVolumeNameFromConfig extracts the actual volume name from the Docker Compose configuration.
// The volumeKey is the logical name (e.g., "production_postgres_data").
// Returns the actual Docker volume name (e.g., "ghostwriter_production_postgres_data").
func (this *DockerInterface) GetVolumeNameFromConfig(volumeKey string) (string, error) {
	volumePath, err := yaml.PathString(fmt.Sprintf("$.volumes.%s.name", volumeKey))
	if err != nil {
		return "", fmt.Errorf("failed to create yaml path: %w", err)
	}

	config, err := this.RunCmdWithOutput("compose", "-f", this.ComposeFile, "config")
	if err != nil {
		return "", fmt.Errorf("failed to get compose config: %w", err)
	}

	var volumeName string
	err = volumePath.Read(strings.NewReader(config), &volumeName)
	if err != nil {
		// Volume might not be explicitly named, try to construct it
		projectName := this.GetComposeProjectName()
		volumeName = fmt.Sprintf("%s_%s", projectName, volumeKey)
	}

	return volumeName, nil
}

// VerifyVolumeExists checks if a Docker volume with the given name exists.
func (this *DockerInterface) VerifyVolumeExists(volumeName string) bool {
	err := this.RunCmd("volume", "inspect", volumeName)
	return err == nil
}

// ListVolumes returns a list of Docker volumes matching the given name filter.
// The filter can be a simple string that will be matched as a prefix.
func (this *DockerInterface) ListVolumes(nameFilter string) ([]string, error) {
	out, err := this.RunCmdWithOutput("volume", "ls", "--format", "{{.Name}}")
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	var matchingVolumes []string
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, nameFilter) {
			matchingVolumes = append(matchingVolumes, line)
		}
	}

	return matchingVolumes, nil
}

// CopyVolume copies data from sourceVol to destVol using a temporary Alpine container.
// This is useful for migrating data between volumes with different names.
func (this *DockerInterface) CopyVolume(sourceVol, destVol string) error {
	// Verify source volume exists
	if !this.VerifyVolumeExists(sourceVol) {
		return fmt.Errorf("source volume does not exist: %s", sourceVol)
	}

	// Create destination volume if it doesn't exist
	if !this.VerifyVolumeExists(destVol) {
		if err := this.RunCmd("volume", "create", destVol); err != nil {
			return fmt.Errorf("failed to create destination volume: %w", err)
		}
	}

	// Use Alpine container to copy data
	// Pattern from restore.go - mount both volumes and use cp -a to preserve permissions
	fmt.Printf("    Copying %s → %s (this may take several minutes)...\n", sourceVol, destVol)

	err := this.RunCmd("run", "--rm",
		"-v", fmt.Sprintf("%s:/source:ro", sourceVol),
		"-v", fmt.Sprintf("%s:/dest", destVol),
		"alpine",
		"sh", "-c",
		"cp -a /source/. /dest/")

	if err != nil {
		return fmt.Errorf("failed to copy volume data: %w", err)
	}

	return nil
}

// VerifyVolumeCopy compares file counts between source and destination volumes.
// Returns the file count in each volume and any error encountered.
func (this *DockerInterface) VerifyVolumeCopy(sourceVol, destVol string) (int, int, error) {
	// Count files in source volume
	sourceOut, err := this.RunCmdWithOutput("run", "--rm",
		"-v", fmt.Sprintf("%s:/data:ro", sourceVol),
		"alpine",
		"sh", "-c",
		"find /data -type f 2>/dev/null | wc -l")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count source files: %w", err)
	}

	// Count files in destination volume
	destOut, err := this.RunCmdWithOutput("run", "--rm",
		"-v", fmt.Sprintf("%s:/data:ro", destVol),
		"alpine",
		"sh", "-c",
		"find /data -type f 2>/dev/null | wc -l")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count destination files: %w", err)
	}

	var sourceCount, destCount int
	fmt.Sscanf(strings.TrimSpace(sourceOut), "%d", &sourceCount)
	fmt.Sscanf(strings.TrimSpace(destOut), "%d", &destCount)

	return sourceCount, destCount, nil
}

// BackupMediaFiles executes the "docker compose" command to back up the media files
// to a tar.gz archive in the postgres_data_backups volume
func (this *DockerInterface) BackupMediaFiles() error {
	// Determine the volume names based on the environment
	var dataVolume, backupVolume string
	if this.UseDevInfra {
		dataVolume = "ghostwriter_local_data"
		backupVolume = "ghostwriter_local_postgres_data_backups"
	} else if this.ManageComposeFile {
		// ModeProd uses ghostwriter_sys prefix
		dataVolume = "ghostwriter_sys_production_data"
		backupVolume = "ghostwriter_sys_production_postgres_data_backups"
	} else {
		// ModeLocalProd uses ghostwriter prefix
		dataVolume = "ghostwriter_production_data"
		backupVolume = "ghostwriter_production_postgres_data_backups"
	}

	// Generate timestamp for backup filename
	timestamp := time.Now().Format("2006_01_02T15_04_05")
	backupFilename := fmt.Sprintf("media_backup_%s.tar.gz", timestamp)

	fmt.Printf("[+] Running `%s` to back up media files from %s...\n", this.command, dataVolume)

	// Create a tar.gz archive of the media volume and store it in the backups volume
	// We use the postgres container because it has access to both volumes
	err := this.RunComposeCmd("run", "--rm",
		"-v", fmt.Sprintf("%s:/source:ro", dataVolume),
		"-v", fmt.Sprintf("%s:/backups", backupVolume),
		"postgres",
		"sh", "-c",
		fmt.Sprintf("tar czf /backups/%s -C /source .", backupFilename))
	if err != nil {
		return fmt.Errorf("failed to back up media files: %w", err)
	}

	fmt.Printf("[+] Media backup created: %s\n", backupFilename)
	return nil
}
