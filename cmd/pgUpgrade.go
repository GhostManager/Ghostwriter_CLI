package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	yaml "github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var pgUpgradeCmd = &cobra.Command{
	Use:   "pg-upgrade",
	Short: "Upgrades the PostgreSQL database",
	Long: `Upgrades the PostgreSQL version.

The production environment is targeted by default. Use the "--mode" argument to upgrade a development environment.
`,
	Run: pgUpgrade,
}

func init() {
	rootCmd.AddCommand(pgUpgradeCmd)
}

func pgUpgrade(cmd *cobra.Command, args []string) {
	dockerInterface := internal.GetDockerInterface(mode)
	dockerInterface.Env.Save()
	interfix := ""
	if dockerInterface.UseDevInfra {
		interfix = "local"
	} else {
		interfix = "production"
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Upgrading PostgreSQL data; it is highly recommended that you make a backup before doing this!\n")
	fmt.Print("Press enter to continue, or Ctrl+C to cancel\n")
	reader.ReadString('\n')

	err := dockerInterface.Down(nil)
	if err != nil {
		log.Fatalf("Error trying to bring down the containers with %s: %v\n", dockerInterface.ComposeFile, err)
	}

	volumeName, networkName := getVolumenAndNetworkName(dockerInterface, interfix)

	fmt.Println("[+] Building Postgres container")
	err = dockerInterface.RunComposeCmd("build", "postgres")
	if err != nil {
		log.Fatalf("Error building postgres container with %s: %v\n", dockerInterface.ComposeFile, err)
	}

	fmt.Println("[+] Getting versions")
	serverVersion := postgresVersionInstalled(dockerInterface)
	dataVersion := postgresVersionForData(dockerInterface)
	if serverVersion == dataVersion {
		fmt.Println("No PostgreSQL upgrade needed")
		return
	}
	fmt.Printf("Upgrading PostgreSQL data from %d to %d\n", dataVersion, serverVersion)

	fmt.Println("[+] Starting old Postgres database")
	err = dockerInterface.RunCmd("run", "-d", "--rm",
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
	err = dockerInterface.RunComposeCmd("run", "-T", "--rm",
		"postgres",
		"bash", "-o", "pipefail", "-euc",
		`source /usr/local/bin/_sourced/constants.sh; PGPASSWORD="${POSTGRES_PASSWORD}" pg_dump -h ghostwriter_postgres_upgrade -U "${POSTGRES_USER}" "${POSTGRES_DB}" | gzip > "${BACKUP_DIR_PATH}/_ghostwriter_postgres_upgrade.sql.gz"`,
	)
	if err != nil {
		fmt.Println("[+] Stopping old Postgres server")
		stopErr := dockerInterface.RunCmd("stop", "ghostwriter_postgres_upgrade")
		if stopErr != nil {
			log.Printf("Could not stop old postgres server: %v\n", err)
		}
		log.Fatalf("Could not run backup: %v\n", err)
	}

	fmt.Println("[+] Stopping old Postgres server")
	err = dockerInterface.RunCmd("stop", "ghostwriter_postgres_upgrade")
	if err != nil {
		log.Fatalf("Could not stop old postgres server: %v\n", err)
	}

	// Wait for volume to release
	time.Sleep(2 * time.Second)

	fmt.Println("[+] Removing old Postgres volume")
	err = dockerInterface.RunCmd("volume", "rm", volumeName)
	if err != nil {
		log.Fatalf("Could not delete old postgres db volume: %v\n", err)
	}

	fmt.Println("[+] Starting new Postgres container")
	err = dockerInterface.RunComposeCmd("up", "-d", "postgres")
	if err != nil {
		log.Fatalf("Could not start new postgresql database: %v\n", err)
	}
	// Wait for it to start
	time.Sleep(10 * time.Second)

	fmt.Println("[+] Restoring data")
	err = dockerInterface.RunComposeCmd("run", "-T", "--rm",
		"postgres",
		"restore",
		"_ghostwriter_postgres_upgrade.sql.gz",
	)
	if err != nil {
		log.Fatalf("Could not start new postgresql database: %v\n", err)
	}

	fmt.Println("[+] All done")
}

func getVolumenAndNetworkName(dockerInterface *internal.DockerInterface, interfix string) (string, string) {
	// Docker returns JSON output, but since YAML is a superset, we can use it to get fields.
	volumePath, err := yaml.PathString(fmt.Sprintf("$.volumes.%s_postgres_data.name", interfix))
	if err != nil {
		log.Fatalf("Could not parse volume path. This is a bug.")
	}
	networkPath, err := yaml.PathString("$.networks.default.name")
	if err != nil {
		log.Fatalf("Could not parse network path. This is a bug.")
	}

	config, err := dockerInterface.RunCmdWithOutput("compose", "-f", dockerInterface.ComposeFile, "config")
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

func postgresVersionInstalled(dockerInterface *internal.DockerInterface) int {
	out, err := dockerInterface.RunCmdWithOutput("compose", "-f", dockerInterface.ComposeFile, "run", "--rm", "postgres", "psql", "--version")
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

func postgresVersionForData(dockerInterface *internal.DockerInterface) int {
	out, err := dockerInterface.RunCmdWithOutput("compose", "-f", dockerInterface.ComposeFile, "run", "--rm", "postgres", "cat", "/var/lib/postgresql/data/PG_VERSION")
	if err != nil {
		log.Fatalf("Error trying to get postgresql data version: %v\n", err)
	}
	majorVersion, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		log.Fatalf("Error trying to parse postgresql data version string %v: %v\n", out, err)
	}
	return majorVersion
}
