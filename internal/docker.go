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

var (
	prodImages = []string{
		"ghostwriter_production_django", "ghostwriter_production_nginx",
		"ghostwriter_production_redis", "ghostwriter_production_postgres",
		"ghostwriter_production_graphql", "ghostwriter_production_queue",
	}
	devImages = []string{
		"ghostwriter_local_django", "ghostwriter_local_redis,",
		"ghostwriter_local_postgres", "ghostwriter_local_graphql",
		"ghostwriter_local_queue",
	}
)

// Execute the ``docker-compose`` commands for a first-time installation with
// the specified YAML file (``yaml`` parameter).
func RunDockerComposeInstall(yaml string) {
	fmt.Printf("[+] Running `docker-compose` for first-time installation with %s...\n", yaml)
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
func RunDockerComposeUpgrade(yaml string) {
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
}

// Execute the ``docker-compose`` commands to restart the environment with
// the specified YAML file (``yaml`` parameter).
func RunDockerComposeRestart(yaml string) {
	fmt.Printf("[+] Running `docker-compose` to restart containers with %s...\n", yaml)
	startErr := RunCmd("docker-compose", []string{"-f", yaml, "start"})
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

// Execute the ``docker-compose`` commands to stop all services in the environment with
// the specified YAML file (``yaml`` parameter).
func RunDockerComposeStop(yaml string) {
	fmt.Printf("[+] Running `docker-compose` to stop services with %s...\n", yaml)
	stopErr := RunCmd("docker-compose", []string{"-f", yaml, "down"})
	if stopErr != nil {
		log.Fatalf("Error trying to stop services with %s: %v\n", yaml, stopErr)
	}
}

// Fetch logs from the the container with the specified ``name`` label (``containerName`` parameter).
func FetchLogs(containerName string) []string {
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
			if container.Labels["name"] == containerName {
				reader, err := cli.ContainerLogs(context.Background(), container.ID, types.ContainerLogsOptions{
					ShowStdout: true,
					ShowStderr: true,
					Tail:       "500",
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
	} else {
		fmt.Println("Failed to find that container")
	}
	return logs
}

// Determine if the container with the specified ``name`` label (``containerName`` parameter) is running.
func isServiceRunning(containerName string) bool {
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
			if container.Labels["name"] == strings.ToLower(containerName) {
				return true
			}
		}
	}
	return false
}

// Determine if the Django application has completed startup based on
// the "Application startup complete" log message.
func isDjangoStarted() bool {
	expectedString := "Application startup complete"
	logs := FetchLogs("ghostwriter_django")
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
				time.Sleep(1 * time.Second)
				counter++
			}
		}
	}
}

// Determine if the container with the specified ``name`` label (``containerName`` parameter) is running.
func GetRunning() {
	fmt.Println("[+] Collecting list of running Ghostwriter containers...")
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
				fmt.Printf("* %s\n", container.Image)
			}
		}
	}
}

// Run Ghostwriter's unit and integration tests via ``docker-compose``.
// The tests are run in the development environment and assume certain values
// will be set for test conditions, so the .env file is temporarily adjusted
// during the test run.
func RunGhostwriterTests() {
	fmt.Println("[+] Running Ghostwriter's unit and integration tests...")
	// Save the current env values we're about to change
	currentActionSecret := ghostEnv.Get("HASURA_GRAPHQL_ACTION_SECRET")
	currentSettignsModule := ghostEnv.Get("DJANGO_SETTINGS_MODULE")
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
	ghostEnv.Set("DJANGO_SETTINGS_MODULE", currentSettignsModule)
	WriteGhostwriterEnvironmentVariables()
}
