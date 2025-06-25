package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	yaml "github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var pgUpgradeCmd = &cobra.Command{
	Use:   "pg-upgrade",
	Short: "Upgrades the PostgreSQL database",
	Long: `Upgrades the PostgreSQL version. A production
environment is installed by default. Use the "--dev" flag to install a development environment.
`,
	Run: pgUpgrade,
}

func init() {
	rootCmd.AddCommand(pgUpgradeCmd)
}

func pgUpgrade(cmd *cobra.Command, args []string) {
	docker.EvaluateDockerComposeStatus()
	yaml := ""
	interfix := ""
	if dev {
		docker.SetDevMode()
		yaml = "local.yml"
		interfix = "local"
	} else {
		docker.SetProductionMode()
		yaml = "production.yml"
		interfix = "production"
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Upgrading PostgreSQL data; it is highly recommended that you make a backup before doing this!\n")
	fmt.Print("Press enter to continue, or Ctrl+C to cancel\n")
	reader.ReadString('\n')

	docker.RunDockerComposeDown(yaml, false)

	volumeName, networkName := getVolumenAndNetworkName(yaml, interfix)

	fmt.Println("[+] Building Postgres container")
	docker.RunCmd("docker", []string{"-f", yaml, "build", "postgres"})

	fmt.Println("[+] Getting versions")
	serverVersion := docker.PostgresVersionInstalled(yaml)
	dataVersion := docker.PostgresVersionForData(yaml)
	if serverVersion == dataVersion {
		fmt.Println("No PostgreSQL upgrade needed")
		return
	}
	fmt.Printf("Upgrading PostgreSQL data from %d to %d\n", dataVersion, serverVersion)

	fmt.Println("[+] Starting old Postgres database")
	err := docker.RunRawCmd("docker", "run", "-d", "--rm",
		"--name", "ghostwriter_postgres_upgrade",
		"--volume", fmt.Sprintf("%s:/var/lib/postgresql/data/", volumeName),
		"--network", networkName,
		fmt.Sprintf("postgres:%d", dataVersion),
	)
	if err != nil {
		log.Fatalf("Could not start old Postgres server: %v\n", err)
	}

	// Wait for it to start
	time.Sleep(10 * time.Second)

	// Run the backup
	fmt.Println("[+] Backing up data")
	err = docker.RunCmd("docker", []string{"-f", yaml, "run", "-T", "--rm",
		"postgres",
		"bash", "-o", "pipefail", "-euc",
		`source /usr/local/bin/_sourced/constants.sh; PGPASSWORD="${POSTGRES_PASSWORD}" pg_dump -h ghostwriter_postgres_upgrade -U "${POSTGRES_USER}" "${POSTGRES_DB}" | gzip > "${BACKUP_DIR_PATH}/_ghostwriter_postgres_upgrade.sql.gz"`,
	})
	if err != nil {
		fmt.Println("[+] Stopping old Postgres server")
		stopErr := docker.RunRawCmd("docker", "stop", "ghostwriter_postgres_upgrade")
		if stopErr != nil {
			log.Printf("Could not stop old postgres server: %v\n", err)
		}
		log.Fatalf("Could not run backup: %v\n", err)
	}

	fmt.Println("[+] Stopping old Postgres server")
	err = docker.RunRawCmd("docker", "stop", "ghostwriter_postgres_upgrade")
	if err != nil {
		log.Fatalf("Could not stop old postgres server: %v\n", err)
	}

	// Wait for volume to release
	time.Sleep(2 * time.Second)

	fmt.Println("[+] Removing old Postgres volume")
	err = docker.RunRawCmd("docker", "volume", "rm", volumeName)
	if err != nil {
		log.Fatalf("Could not delete old postgres db volume: %v\n", err)
	}

	fmt.Println("[+] Starting new Postgres container")
	err = docker.RunCmd("docker", []string{"-f", yaml, "up", "-d", "postgres"})
	if err != nil {
		log.Fatalf("Could not start new postgresql database: %v\n", err)
	}
	// Wait for it to start
	time.Sleep(10 * time.Second)

	fmt.Println("[+] Restoring data")
	err = docker.RunCmd("docker", []string{"-f", yaml, "run", "-T", "--rm",
		"postgres",
		"restore",
		"_ghostwriter_postgres_upgrade.sql.gz",
	})
	if err != nil {
		log.Fatalf("Could not start new postgresql database: %v\n", err)
	}

	fmt.Println("[+] All done")
}

func getVolumenAndNetworkName(path string, interfix string) (string, string) {
	volumePath, err := yaml.PathString(fmt.Sprintf("$.volumes.%s_postgres_data.name", interfix))
	if err != nil {
		log.Fatalf("Could not parse volume path. This is a bug.")
	}
	networkPath, err := yaml.PathString("$.networks.default.name")
	if err != nil {
		log.Fatalf("Could not parse network path. This is a bug.")
	}

	config, err := docker.RunBasicCmd("docker", []string{"compose", "-f", path, "config"})
	if err != nil {
		log.Fatalf("Could not get docker config: %s\n", err)
	}

	var volume string
	err = volumePath.Read(strings.NewReader(config), &volume)
	if err != nil {
		log.Fatalf("Could not get volume path: %s\n", err)
	}

	var network string
	err = networkPath.Read(strings.NewReader(config), &network)
	if err != nil {
		log.Fatalf("Could not get network path: %s\n", err)
	}

	return volume, network
}
