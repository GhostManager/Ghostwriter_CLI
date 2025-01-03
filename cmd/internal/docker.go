package internal

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Vars for tracking the list of Ghostwriter images
// Used for filtering the list of containers returned by the Docker client
var (
	prodImages = []string{
		"ghostwriter_production_django", "ghostwriter_production_nginx",
		"ghostwriter_production_redis", "ghostwriter_production_postgres",
		"ghostwriter_production_graphql", "ghostwriter_production_queue",
	}
	devImages = []string{
		"ghostwriter_local_django", "ghostwriter_local_redis",
		"ghostwriter_local_postgres", "ghostwriter_local_graphql",
		"ghostwriter_local_queue",
	}
	// Default root command for Docker commands
	dockerCmd = "docker"
)

// Container is a custom type for storing container information similar to output from “docker containers ls“.
type Container struct {
	ID     string
	Image  string
	Status string
	Ports  []types.Port
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

// EvaluateDockerComposeStatus determines if the host has the “docker compose“ plugin or the “docker compose“
// script installed and set the global `dockerCmd` variable.
func EvaluateDockerComposeStatus() error {
	fmt.Println("[+] Checking the status of Docker and the Compose plugin...")
	// Check for ``docker`` first because it's required for everything to come
	dockerExists := CheckPath("docker")
	if !dockerExists {
		log.Fatalln("Docker is not installed on this system, so please install Docker and try again")
	}

	// Check if the Docker Engine is running
	_, engineErr := RunBasicCmd("docker", []string{"info"})
	if engineErr != nil {
		log.Fatalln("Docker is installed on this system, but the daemon is not running")
	}

	// Check for the ``compose`` plugin as our first choice
	_, composeErr := RunBasicCmd("docker", []string{"compose", "version"})
	if composeErr != nil {
		fmt.Println("[+] The `compose` is not installed, so we'll try the deprecated `docker-compose` script")
		composeScriptExists := CheckPath("docker-compose")
		if composeScriptExists {
			fmt.Println("[+] The `docker-compose` script is installed, so we'll use that instead")
			dockerCmd = "docker-compose"
		} else {
			fmt.Println("[+] The `docker-compose` script is also not installed or in the PATH")
			log.Fatalln("Docker Compose is not installed, so please install it and try again: https://docs.docker.com/compose/install/")
		}
	}

	// Bail out if we're not in the same directory as the YAML files
	// Otherwise, we'll get a confusing error message from the `compose` plugin
	if !FileExists(filepath.Join(GetCwdFromExe(), "local.yml")) || !FileExists(filepath.Join(GetCwdFromExe(), "production.yml")) {
		log.Fatalln("Ghostwriter CLI must be run in the same directory as the `local.yml` and `production.yml` files")
	}

	return nil
}

// RunDockerComposeInstall executes the “docker compose“ commands for a first-time installation with
// the specified YAML file (“yaml“ parameter).
func RunDockerComposeInstall(yaml string) {
	buildErr := RunCmd(dockerCmd, []string{"-f", yaml, "build"})
	if buildErr != nil {
		log.Fatalf("Error trying to build with %s: %v\n", yaml, buildErr)
	}
	upErr := RunCmd(dockerCmd, []string{"-f", yaml, "up", "-d"})
	if upErr != nil {
		log.Fatalf("Error trying to bring up environment with %s: %v\n", yaml, upErr)
	}
	// Must wait for Django to complete db migrations before seeding the database
	for {
		if waitForDjango() {
			fmt.Println("[+] Proceeding with Django database setup...")
			seedErr := RunCmd(dockerCmd, []string{"-f", yaml, "run", "--rm", "django", "/seed_data"})
			if seedErr != nil {
				log.Fatalf("Error trying to seed the database: %v\n", seedErr)
			}
			fmt.Println("[+] Proceeding with Django superuser creation...")
			userErr := RunCmd(
				dockerCmd, []string{"-f", yaml, "run", "--rm", "django", "python",
					"manage.py", "createsuperuser", "--noinput", "--role", "admin"},
			)
			// This may fail if the user has already created a superuser, so we don't exit
			if userErr != nil {
				log.Printf("Error trying to create a superuser: %v\n", userErr)
				log.Println("Error may occur if you've run `install` before or made a superuser manually")
			}
			break
		}
	}
	// Restart Hasura to ensure metadata matches post-migrations and seeding
	restartErr := RunCmd(dockerCmd, []string{"-f", yaml, "restart", "graphql_engine"})
	if restartErr != nil {
		fmt.Printf("[-] Error trying to restart the `graphql_engine` service: %v\n", restartErr)
	}
	fmt.Println("[+] Ghostwriter is ready to go!")
	fmt.Printf("[+] You can login as `%s` with this password: %s\n", ghostEnv.GetString("django_superuser_username"), ghostEnv.GetString("django_superuser_password"))
	fmt.Println("[+] You can get your admin password by running: ghostwriter-cli config get admin_password")
}

// RunDockerComposeUpgrade executes the “docker compose“ commands for re-building or upgrading an
// installation with the specified YAML file (“yaml“ parameter).
func RunDockerComposeUpgrade(yaml string, skipseed bool) {
	fmt.Printf("[+] Running `%s` commands to build containers with %s...\n", dockerCmd, yaml)
	downErr := RunCmd(dockerCmd, []string{"-f", yaml, "down"})
	if downErr != nil {
		log.Fatalf("Error trying to bring down any running containers with %s: %v\n", yaml, downErr)
	}
	buildErr := RunCmd(dockerCmd, []string{"-f", yaml, "build"})
	if buildErr != nil {
		log.Fatalf("Error trying to build with %s: %v\n", yaml, buildErr)
	}
	upErr := RunCmd(dockerCmd, []string{"-f", yaml, "up", "-d"})
	if upErr != nil {
		log.Fatalf("Error trying to bring up environment with %s: %v\n", yaml, upErr)
	}
	if !skipseed {
		// Must wait for Django to complete any potential db migrations before re-seeding the database
		for {
			if waitForDjango() {
				fmt.Println("[+] Re-seeding database in case initial values were added or adjusted...")
				seedErr := RunCmd(dockerCmd, []string{"-f", yaml, "run", "--rm", "django", "/seed_data"})
				if seedErr != nil {
					log.Fatalf("Error trying to seed the database: %v\n", seedErr)
				}
				break
			}
		}
	} else {
		fmt.Println("[+] The `--skip-seed` flag was set, so skipped database seeding...")
	}
	fmt.Println("[+] All containers have been built!")
}

// RunDockerComposeStart executes the “docker compose“ commands to start the environment with
// the specified YAML file (“yaml“ parameter).
func RunDockerComposeStart(yaml string) {
	fmt.Printf("[+] Running `%s` to restart containers with %s...\n", dockerCmd, yaml)
	startErr := RunCmd(dockerCmd, []string{"-f", yaml, "start"})
	if startErr != nil {
		log.Fatalf("Error trying to restart the containers with %s: %v\n", yaml, startErr)
	}
}

// RunDockerComposeStop executes the “docker compose“ commands to stop all services in the environment with
// the specified YAML file (“yaml“ parameter).
func RunDockerComposeStop(yaml string) {
	fmt.Printf("[+] Running `%s` to stop services with %s...\n", dockerCmd, yaml)
	stopErr := RunCmd(dockerCmd, []string{"-f", yaml, "stop"})
	if stopErr != nil {
		log.Fatalf("Error trying to stop services with %s: %v\n", yaml, stopErr)
	}
}

// RunDockerComposeRestart executes the “docker compose“ commands to restart the environment with
// the specified YAML file (“yaml“ parameter).
func RunDockerComposeRestart(yaml string) {
	fmt.Printf("[+] Running `%s` to restart containers with %s...\n", dockerCmd, yaml)
	startErr := RunCmd(dockerCmd, []string{"-f", yaml, "restart"})
	if startErr != nil {
		log.Fatalf("Error trying to restart the containers with %s: %v\n", yaml, startErr)
	}
}

// RunDockerComposeUp executes the “docker compose“ commands to bring up the environment with
// the specified YAML file (“yaml“ parameter).
func RunDockerComposeUp(yaml string) {
	fmt.Printf("[+] Running `%s` to bring up the containers with %s...\n", dockerCmd, yaml)
	upErr := RunCmd(dockerCmd, []string{"-f", yaml, "up", "-d"})
	if upErr != nil {
		log.Fatalf("Error trying to bring up the containers with %s: %v\n", yaml, upErr)
	}
}

// RunDockerComposeDown executes the “docker compose“ commands to bring down the environment with
// the specified YAML file (“yaml“ parameter).
func RunDockerComposeDown(yaml string) {
	fmt.Printf("[+] Running `%s` to bring down the containers with %s...\n", dockerCmd, yaml)
	downErr := RunCmd(dockerCmd, []string{"-f", yaml, "down"})
	if downErr != nil {
		log.Fatalf("Error trying to bring down the containers with %s: %v\n", yaml, downErr)
	}
}

// FetchLogs fetches logs from the container with the specified “name“ label (“containerName“ parameter).
func FetchLogs(containerName string, lines string) []string {
	var logs []string
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to get client in logs: %v", err)
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Fatalf("Failed to get container list: %v", err)
	}
	if len(containers) > 0 {
		for _, container := range containers {
			if container.Labels["name"] == containerName || containerName == "all" || container.Labels["name"] == "ghostwriter_"+containerName {
				logs = append(logs, fmt.Sprintf("\n*** Logs for `%s` ***\n\n", container.Labels["name"]))
				reader, err := cli.ContainerLogs(context.Background(), container.ID, types.ContainerLogsOptions{
					ShowStdout: true,
					ShowStderr: true,
					Tail:       lines,
				})
				if err != nil {
					log.Fatalf("Failed to get container logs: %v", err)
				}
				defer reader.Close()
				// Reference: https://medium.com/@dhanushgopinath/reading-docker-container-logs-with-golang-docker-engine-api-702233fac044
				p := make([]byte, 8)
				_, err = reader.Read(p)
				for err == nil {
					content := make([]byte, binary.BigEndian.Uint32(p[4:]))
					reader.Read(content)
					logs = append(logs, string(content))
					_, err = reader.Read(p)
				}
			}
		}

		if len(logs) == 0 {
			logs = append(logs, fmt.Sprintf("\n*** No logs found for requested container '%s' ***\n", containerName))
		}
	} else {
		fmt.Println("Failed to find that container")
	}
	return logs
}

// GetRunning determines if the container with the specified “name“ label (“containerName“ parameter) is running.
func GetRunning() Containers {
	var running Containers

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to get client connection to Docker: %v", err)
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		All: false,
	})
	if err != nil {
		log.Fatalf("Failed to get container list from Docker: %v", err)
	}
	if len(containers) > 0 {
		for _, container := range containers {
			if Contains(devImages, container.Image) || Contains(prodImages, container.Image) {
				running = append(running, Container{
					container.ID, container.Image, container.Status, container.Ports, container.Labels["name"],
				})
			}
		}
	}

	return running
}

// Determine if the container with the specified “name“ label (“containerName“ parameter) is running.
func isServiceRunning(containerName string) bool {
	containers := GetRunning()
	for _, container := range containers {
		if container.Name == strings.ToLower(containerName) {
			return true
		}
	}
	return false
}

// Determine if the Django application has completed startup based on
// the "Application startup complete" log message.
func isDjangoStarted() bool {
	expectedString := "Application startup complete"
	logs := FetchLogs("ghostwriter_django", "500")
	for _, entry := range logs {
		result := strings.Contains(entry, expectedString)
		if result {
			return true
		}
	}
	return false
}

// Check if PostgreSQL is having trouble starting due to a password mismatch.
func isPostgresStarted() bool {
	expectedString := "Password does not match for user"
	logs := FetchLogs("ghostwriter_postgres", "100")
	for _, entry := range logs {
		result := strings.Contains(entry, expectedString)
		if result {
			return true
		}
	}
	return false
}

// Determine if the Ghostwriter application has completed startup
func waitForDjango() bool {
	// Wait for ghostwriter to start running
	fmt.Println("[+] Waiting for Django application startup to complete...")
	counter := 0
	for {
		if !isServiceRunning("ghostwriter_django") {
			fmt.Print("\n")
			log.Fatalf("Django container exited unexpectedly. Check the logs in docker for the ghostwriter_django container")
		}
		if isDjangoStarted() {
			fmt.Print("\n[+] Django application started\n")
			return true
		}
		if isPostgresStarted() {
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

// RunGhostwriterTests runs Ghostwriter's unit and integration tests via “docker compose“.
// The tests are run in the development environment and assume certain values
// will be set for test conditions, so the .env file is temporarily adjusted
// during the test run.
func RunGhostwriterTests() {
	// Save the current env values we're about to change
	currentActionSecret := ghostEnv.Get("HASURA_GRAPHQL_ACTION_SECRET")
	currentSettingsModule := ghostEnv.Get("DJANGO_SETTINGS_MODULE")

	// Change env values for the test conditions
	ghostEnv.Set("HASURA_GRAPHQL_ACTION_SECRET", "changeme")
	ghostEnv.Set("DJANGO_SETTINGS_MODULE", "config.settings.local")
	WriteGhostwriterEnvironmentVariables()

	// Run the unit tests
	testErr := RunCmd(dockerCmd, []string{"-f", "local.yml", "run", "--rm", "django", "python", "manage.py", "test"})
	if testErr != nil {
		log.Fatalf("Error trying to run Ghostwriter's tests: %v\n", testErr)
	}

	// Reset the changed env values
	ghostEnv.Set("HASURA_GRAPHQL_ACTION_SECRET", currentActionSecret)
	ghostEnv.Set("DJANGO_SETTINGS_MODULE", currentSettingsModule)
	WriteGhostwriterEnvironmentVariables()
}

// CheckDockerHealth determines if all containers are running and passing their respective health checks.
func CheckDockerHealth(dev bool) (HealthIssues, error) {
	var found []string
	var imageName string
	var issues HealthIssues

	requiredImages := prodImages
	if dev {
		requiredImages = devImages
	}

	// Check running containers to make sure every necessary container is up
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return issues, err
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		All: false,
	})
	if err != nil {
		return issues, err
	}

	if len(containers) > 0 {
		for _, container := range containers {
			if Contains(devImages, container.Image) || Contains(prodImages, container.Image) {
				found = append(found, container.Image)
			}
		}
		for _, image := range requiredImages {
			if !Contains(found, image) {
				imageName = strings.ToUpper(image[strings.LastIndex(image, "_")+1:])
				issues = append(issues, HealthIssue{"Container", imageName, "Container is not running"})
			}
		}
	} else {
		issues = append(issues, HealthIssue{"Container", "ALL", "No Ghostwriter containers are running"})
	}

	return issues, nil
}

// RunDockerComposeBackup executes the “docker compose“ command to back up the PostgreSQL database in the environment
// from the specified YAML file (“yaml“ parameter).
func RunDockerComposeBackup(yaml string) {
	fmt.Printf("[+] Running `%s` to back up the PostgreSQL database with %s...\n", dockerCmd, yaml)
	backupErr := RunCmd(dockerCmd, []string{"-f", yaml, "run", "--rm", "postgres", "backup"})
	if backupErr != nil {
		log.Fatalf("Error trying to back up the PostgreSQL database with %s: %v\n", yaml, backupErr)
	}
}

// RunDockerComposeBackups executes the “docker compose“ command to list available PostgreSQL database backups in the
// environment from the specified YAML file (“yaml“ parameter).
func RunDockerComposeBackups(yaml string) {
	fmt.Printf("[+] Running `%s` to list avilable PostgreSQL database backup files with %s...\n", dockerCmd, yaml)
	backupErr := RunCmd(dockerCmd, []string{"-f", yaml, "run", "--rm", "postgres", "backups"})
	if backupErr != nil {
		log.Fatalf("Error trying to list backups files with %s: %v\n", yaml, backupErr)
	}
}

// RunDockerComposeRestore executes the “docker compose“ command to restore a PostgreSQL database backup in the
// environment from the specified YAML file (“yaml“ parameter).
func RunDockerComposeRestore(yaml string, restore string) {
	fmt.Printf("[+] Running `%s` to restore the PostgreSQL database backup file %s with %s...\n", dockerCmd, restore, yaml)
	backupErr := RunCmd(dockerCmd, []string{"-f", yaml, "run", "--rm", "postgres", "restore", restore})
	if backupErr != nil {
		log.Fatalf("Error trying to restore %s with %s: %v\n", restore, yaml, backupErr)
	}
}

// Gets the major version number of the PostgreSQL installation
func PostgresVersionInstalled(
	yaml string,
) int {
	out, err := RunBasicCmd("docker", []string{"compose", "-f", yaml, "run", "--rm", "postgres", "psql", "--version"})
	if err != nil {
		log.Fatalf("Error trying to get postgresql server version: %v\n", err)
	}

	match := regexp.MustCompile(`(\d+)\.\d+`).FindStringSubmatch(out)
	if len(match) == 0 {
		log.Fatalf("Could not find version in string %v", out)
	}

	majorVersion, err := strconv.Atoi(match[1])
	if err != nil {
		log.Fatalf("Could not parse installed Postgres version of %v: %v", match[1], err)
	}
	return majorVersion
}

// Gets the major version number of the PostgreSQL data. If different from the installation version, an upgrade is needed.
func PostgresVersionForData(
	yaml string,
) int {
	out, err := RunBasicCmd("docker", []string{"compose", "-f", yaml, "run", "--rm", "postgres", "cat", "/var/lib/postgresql/data/PG_VERSION"})
	if err != nil {
		log.Fatalf("Error trying to get postgresql data version: %v\n", err)
	}
	majorVersion, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		log.Fatalf("Error trying to parse postgresql data version string %v: %v\n", out, err)
	}
	return majorVersion
}
