package cmd

import (
	"fmt"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// containersStopCmd represents the stop command
var containersStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all Ghostwriter services without removing the containers",
	Long: `Stop all Ghostwriter services without removing the containers. This
performs the equivalent of running the "docker compose stop" command.

Production containers are targeted by default. Use the "--dev" flag to
target development containers`,
	Run: containersStop,
}

func init() {
	containersCmd.AddCommand(containersStopCmd)
}

func containersStop(cmd *cobra.Command, args []string) {
	docker.EvaluateDockerComposeStatus()
	if dev {
		fmt.Println("[+] Stopping the development environment")
		docker.RunDockerComposeStop("local.yml")
	} else {
		fmt.Println("[+] Stopping the production environment")
		docker.RunDockerComposeStop("production.yml")
	}
}
