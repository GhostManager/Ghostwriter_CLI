package cmd

import (
	"fmt"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// containersUpCmd represents the up command
var containersUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Build, (re)create, and start all Ghostwriter containers",
	Long: `Build, (re)create, and start all Ghostwriter containers. This
performs the equivilant of running the "docker-compose up" command.

Production containers are targeted by default. Use the "--dev" flag to
target development containers`,
	Run: containersUp,
}

func init() {
	containersCmd.AddCommand(containersUpCmd)
}

func containersUp(cmd *cobra.Command, args []string) {
	if dev {
		fmt.Println("[+] Bringing up the development environment")
		docker.SetDevMode()
		docker.RunDockerComposeUp("local.yml")
	} else {
		fmt.Println("[+] Bringing up the production environment")
		docker.SetProductionMode()
		docker.RunDockerComposeUp("production.yml")
	}
}
