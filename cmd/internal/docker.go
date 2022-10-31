package internal

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"strings"
	"time"
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
)

// Custom type for storing container information similar to output from ``docker containers ls``.
type Container struct {
	ID     string
	Image  string
	Status string
	Ports  []types.Port
	Name   string
}

type Containers []Container

func (c Containers) Len() int {
	return len(c)
}

func (c Containers) Less(i, j int) bool {
	return c[i].Image < c[j].Image
}

func (c Containers) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Execute the ``docker-compose`` commands for a first-time installation with
// the specified YAML file (``yaml`` parameter).
func RunDockerComposeInstall(yaml string) {
	buildErr := RunCmd("docker-compose", []string{"-f", yaml, "build"})
	if buildErr != nil {
		log.Fatalf("Error trying to build with %s: %v\n", yaml, buildErr)
	}
	upErr := RunCmd("docker-compose", []string{"-f", yaml, "up", "-d"})
	if upErr != nil {
		log.Fatalf("Error trying to bring up environment with %s: %v\n", yaml, upErr)
	}
	// Must wait for Django to complete db migrations before seeding the database
	for {
		if waitForDjango() {
			fmt.Println("[+] Proceeding with Django database setup...")
			seedErr := RunCmd("docker-compose", []string{"-f", yaml, "run", "--rm", "django", "/seed_data"})
			if seedErr != nil {
				log.Fatalf("Error trying to seed the database: %v\n", seedErr)
			}
			fmt.Println("[+] Proceeding with Django superuser creation...")
			userErr := RunCmd(
				"docker-compose", []string{"-f", yaml, "run", "--rm", "django", "python",
					"manage.py", "createsuperuser", "--noinput"},
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
	restartErr := RunCmd("docker-compose", []string{"-f", yaml, "restart", "graphql_engine"})
	if restartErr != nil {
		fmt.Printf("[-] Error trying to restart the `graphql_engine` service: %v\n", restartErr)
	}
	fmt.Println("[+] Ghostwriter is ready to go!")
	fmt.Printf("[+] You can login as `%s` with this password: %s\n", ghostEnv.GetString("django_superuser_username"), ghostEnv.GetString("django_superuser_password"))
	fmt.Println("[+] You can get your admin password by running: ghostwriter-cli config get admin_password")
}

// Execute the ``docker-compose`` commands for re-building or upgrading an
// installation with the specified YAML file (``yaml`` parameter).
func RunDockerComposeUpgrade(yaml string, skipseed bool) {
	fmt.Printf("[+] Running `docker-compose` commands to build containers with %s...\n", yaml)
	downErr := RunCmd("docker-compose", []string{"-f", yaml, "down"})
	if downErr != nil {
		log.Fatalf("Error trying to bring down any running containers with %s: %v\n", yaml, downErr)
	}
	buildErr := RunCmd("docker-compose", []string{"-f", yaml, "build"})
	if buildErr != nil {
		log.Fatalf("Error trying to build with %s: %v\n", yaml, buildErr)
	}
	upErr := RunCmd("docker-compose", []string{"-f", yaml, "up", "-d"})
	if upErr != nil {
		log.Fatalf("Error trying to bring up environment with %s: %v\n", yaml, upErr)
	}
	if !skipseed {
		// Must wait for Django to complete any potential db migrations before re-seeding the database
		for {
			if waitForDjango() {
				fmt.Println("[+] Re-seeding database in case initial values were added or adjusted...")
				seedErr := RunCmd("docker-compose", []string{"-f", yaml, "run", "--rm", "django", "/seed_data"})
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

// Execute the ``docker-compose`` commands to start the environment with
// the specified YAML file (``yaml`` parameter).
func RunDockerComposeStart(yaml string) {
	fmt.Printf("[+] Running `docker-compose` to restart containers with %s...\n", yaml)
	startErr := RunCmd("docker-compose", []string{"-f", yaml, "start"})
	if startErr != nil {
		log.Fatalf("Error trying to restart the containers with %s: %v\n", yaml, startErr)
	}
}

// Execute the ``docker-compose`` commands to stop all services in the environment with
// the specified YAML file (``yaml`` parameter).
func RunDockerComposeStop(yaml string) {
	fmt.Printf("[+] Running `docker-compose` to stop services with %s...\n", yaml)
	stopErr := RunCmd("docker-compose", []string{"-f", yaml, "stop"})
	if stopErr != nil {
		log.Fatalf("Error trying to stop services with %s: %v\n", yaml, stopErr)
	}
}

// Execute the ``docker-compose`` commands to restart the environment with
// the specified YAML file (``yaml`` parameter).
func RunDockerComposeRestart(yaml string) {
	fmt.Printf("[+] Running `docker-compose` to restart containers with %s...\n", yaml)
	startErr := RunCmd("docker-compose", []string{"-f", yaml, "restart"})
	if startErr != nil {
		log.Fatalf("Error trying to restart the containers with %s: %v\n", yaml, startErr)
	}
}

// Execute the ``docker-compose`` commands to bring up the environment with
// the specified YAML file (``yaml`` parameter).
func RunDockerComposeUp(yaml string) {
	fmt.Printf("[+] Running `docker-compose` to bring up the containers with %s...\n", yaml)
	upErr := RunCmd("docker-compose", []string{"-f", yaml, "up", "-d"})
	if upErr != nil {
		log.Fatalf("Error trying to bring up the containers with %s: %v\n", yaml, upErr)
	}
}

// Execute the ``docker-compose`` commands to bring down the environment with
// the specified YAML file (``yaml`` parameter).
func RunDockerComposeDown(yaml string) {
	fmt.Printf("[+] Running `docker-compose` to bring down the containers with %s...\n", yaml)
	downErr := RunCmd("docker-compose", []string{"-f", yaml, "down"})
	if downErr != nil {
		log.Fatalf("Error trying to bring down the containers with %s: %v\n", yaml, downErr)
	}
}

// Fetch logs from the the container with the specified ``name`` label (``containerName`` parameter).
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

// Determine if the container with the specified ``name`` label (``containerName`` parameter) is running.
func GetRunning() Containers {
	var running Containers

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to get client connection to Docker: %v", err)
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
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

// Determine if the container with the specified ``name`` label (``containerName`` parameter) is running.
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
	for {
		if isServiceRunning("ghostwriter_django") {
			fmt.Println("[+] Waiting for Django application startup to complete...")
			counter := 1
			for {
				fmt.Printf("\r%s", strings.Repeat(".", counter))
				if isDjangoStarted() {
					fmt.Print("\n[+] Django application started\n")
					return true
				}
				if isPostgresStarted() {
					log.Fatalf("\nPostgreSQL cannot start because of a password mismatch. Please read: https://www.ghostwriter.wiki/getting-help/faq#ghostwriter-cli-reports-an-issue-with-postgresql")
				}
				time.Sleep(1 * time.Second)
				counter++
			}
		}
	}
}

// Run Ghostwriter's unit and integration tests via ``docker-compose``.
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
	testErr := RunCmd("docker-compose", []string{"-f", "local.yml", "run", "--rm", "django", "python", "manage.py", "test"})
	if testErr != nil {
		log.Fatalf("Error trying to run Ghostwriter's tests: %v\n", testErr)
	}

	// Reset the changed env values
	ghostEnv.Set("HASURA_GRAPHQL_ACTION_SECRET", currentActionSecret)
	ghostEnv.Set("DJANGO_SETTINGS_MODULE", currentSettingsModule)
	WriteGhostwriterEnvironmentVariables()
}
