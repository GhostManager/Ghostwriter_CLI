package cmd

import (
	"fmt"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

var volumes bool

// containersDownCmd represents the down command
var containersDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Bring down all Ghostwriter services and remove the containers",
	Long: `Bring down all Ghostwriter services and remove the containers. This
performs the equivalent of running the "docker compose down" command.

Production containers are targeted by default. Use the "--dev" flag to
target development containers`,
	Run: containersDown,
}

func init() {
	containersCmd.AddCommand(containersDownCmd)

	containersDownCmd.PersistentFlags().BoolVar(&volumes, "volumes", false, "Delete data volumes when containers come down")
}

func containersDown(cmd *cobra.Command, args []string) {
	docker.EvaluateDockerComposeStatus()
	if dev {
		fmt.Println("[+] Bringing down the development environment")
		docker.RunDockerComposeDown("local.yml", volumes)
	} else {
		fmt.Println("[+] Bringing down the production environment")
		docker.RunDockerComposeDown("production.yml", volumes)
	}
}
