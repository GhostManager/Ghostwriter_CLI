package cmd

import (
	"fmt"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// containersRestartCmd represents the restart command
var containersRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart all stopped and running Ghostwriter services",
	Long: `Restart all stopped and running Ghostwriter services. This performs
the equivalent of running the "docker-compose restart" command.

Production containers are targeted by default. Use the "--dev" flag to
target development containers`,
	Run: containersRestart,
}

func init() {
	containersCmd.AddCommand(containersRestartCmd)
}

func containersRestart(cmd *cobra.Command, args []string) {
	if dev {
		fmt.Println("[+] Restarting the development environment")
		docker.RunDockerComposeRestart("local.yml")
	} else {
		fmt.Println("[+] Restarting the production environment")
		docker.RunDockerComposeRestart("production.yml")
	}
}
