package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs and sets up Ghostwriter",
	Long: `Installs and sets up Ghostwriter. By default, Ghostwriter will download and
install the latest version to an application data directory - use the "--mode" option to use a
source checkout instead.

The command performs the following steps:

* Sets up the default server configuration
* Generates TLS certificates for the server
* Fetches or builds the Docker containers
* Creates a default admin user with a randomly generated password

Running after initial installation will keep the existing configuration but fetch a new version
(for --mode=prod) or rebuild the containers (for --mode=local-*)
`,
	Run: installGhostwriter,
}

var installVersion string

func init() {
	installCmd.PersistentFlags().StringVar(
		&installVersion,
		"version",
		"",
		"Version to install. Defaults to the latest tagged release. Ignored for --mode=local-*. NOTE: downgrading is not supported.",
	)
	rootCmd.AddCommand(installCmd)
}

func fetchAndWriteComposeFile(mode internal.DockerMode, version string) error {
	dir := docker.GetDockerDirFromMode(mode)
	file := "docker-compose.yml"

	fmt.Println("[+] Downloading docker-compose.yml")

	var url string
	if version == "" {
		url = "https://github.com/GhostManager/Ghostwriter/releases/latest/download/gw-cli.yml"
	} else {
		// Validate version format to prevent path traversal or malicious content
		versionPattern := regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
		if !versionPattern.MatchString(version) {
			return fmt.Errorf("invalid version format '%s': expected format like v1.2.3", version)
		}
		fmt.Printf("[+] Fetching docker-compose.yml for version %s\n", version)
		url = "https://github.com/GhostManager/Ghostwriter/releases/download/" + version + "/gw-cli.yml"
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Could not create request for gw-cli.yml: %w", err)
	}
	req.Header.Set("User-Agent", "Ghostwriter-CLI")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Could not get gw-cli.yml from github: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		if res.StatusCode == 404 {
			return fmt.Errorf("Could not get gw-cli.yml from github: status code %d\n(Ghostwriter-CLI cannot install versions of Ghostwriter older than v6.2.3 in `--mode=prod`. If you're trying to install a version later than that, try updating Ghostwriter-CLI)", res.StatusCode)
		}
		return fmt.Errorf("Could not get gw-cli.yml from github: status code %d", res.StatusCode)
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("Could not get gw-cli.yml from github: %w", err)
	}

	err = os.WriteFile(
		filepath.Join(dir, file),
		buf,
		0644,
	)

	if err != nil {
		return fmt.Errorf("Could not write docker-compose.yml file: %w", err)
	}
	return nil
}

// Performs common setup
func updateContainers(dockerInterface docker.DockerInterface) error {
	var err error
	if dockerInterface.ManageComposeFile {
		fmt.Println("[+] Pulling containers...")
		err = dockerInterface.RunComposeCmd("pull")
		if err != nil {
			return fmt.Errorf("Could not pull containers: %w", err)
		}
	} else {
		fmt.Println("[+] Building containers...")
		err = dockerInterface.RunComposeCmd("build", "--pull")
		if err != nil {
			return fmt.Errorf("Could not build containers: %w", err)
		}
	}

	fmt.Println("[+] Starting containers...")
	err = dockerInterface.Up()
	if err != nil {
		return fmt.Errorf("Could not start containers: %w", err)
	}

	fmt.Println("[+] Waiting for Django to be ready...")
	dockerInterface.WaitForDjango()

	fmt.Println("[+] Migrating database...")
	err = dockerInterface.RunDjangoManageCommand("migrate")
	if err != nil {
		return fmt.Errorf("Could not migrate database: %w", err)
	}

	fmt.Println("[+] Seeding database with initial data...")
	err = dockerInterface.RunComposeCmd("run", "--rm", "django", "/seed_data")
	if err != nil {
		return fmt.Errorf("Could not seed database: %w", err)
	}

	return nil
}

func installGhostwriter(cmd *cobra.Command, args []string) {
	var err error

	if mode == docker.ModeProd {
		// Fetch and write docker-compose.yml file
		err = fetchAndWriteComposeFile(mode, installVersion)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}

	// Get interface
	dockerInterface := docker.GetDockerInterface(mode)
	dockerInterface.Env.Save()
	if dockerInterface.UseDevInfra {
		fmt.Println("[+] Starting development environment installation")
	} else {
		fmt.Println("[+] Starting production environment installation")
		docker.GenerateCertificatePackage(dockerInterface.Dir)
		docker.PrepareSettingsDirectory(dockerInterface.Dir)
	}

	err = updateContainers(*dockerInterface)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	fmt.Println("[+] Proceeding with Django superuser creation...")
	userErr := dockerInterface.RunDjangoManageCommand("createsuperuser", "--noinput", "--role", "admin")
	// This may fail if the user has already created a superuser, so we don't exit
	if userErr != nil {
		log.Printf("Error trying to create a superuser: %v\n", userErr)
		log.Println("Error may occur if you've run `install` before or made a superuser manually")
	}

	fmt.Println("[+] Ghostwriter is ready to go!")
	fmt.Printf("[+] You can log in as `%s` with this password: %s\n", dockerInterface.Env.Get("django_superuser_username"), dockerInterface.Env.Get("django_superuser_password"))
	fmt.Println("[+] You can get your admin password by running: ghostwriter-cli config get admin_password")
}
