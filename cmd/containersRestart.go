package cmd

import (
	"fmt"
	"log"

	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// containersRestartCmd represents the restart command
var containersRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart all stopped and running Ghostwriter services",
	Long: `Restart all stopped and running Ghostwriter services. This performs
the equivalent of running the "docker compose restart" command.

Production containers are targeted by default. Use the "--mode" argument to
target development containers`,
	Run: containersRestart,
}

func init() {
	containersCmd.AddCommand(containersRestartCmd)
}

func containersRestart(cmd *cobra.Command, args []string) {
	dockerInterface := internal.GetDockerInterface(mode)
	if dockerInterface.UseDevInfra {
		fmt.Println("[+] Restarting the development environment")
	} else {
		fmt.Println("[+] Restarting the production environment")
	}

	fmt.Printf("[+] Restarting containers with %s...\n", dockerInterface.ComposeFile)
	startErr := dockerInterface.RunComposeCmd("restart")
	if startErr != nil {
		log.Fatalf("Error trying to restart the containers with %s: %v\n", dockerInterface.ComposeFile, startErr)
	}
}
