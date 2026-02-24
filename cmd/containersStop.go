package cmd

import (
	"fmt"
	"log"

	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// containersStopCmd represents the stop command
var containersStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all Ghostwriter services without removing the containers",
	Long: `Stop all Ghostwriter services without removing the containers. This
performs the equivalent of running the "docker compose stop" command.

Production containers are targeted by default. Use the "--mode" argument to
target development containers`,
	Run: containersStop,
}

func init() {
	containersCmd.AddCommand(containersStopCmd)
}

func containersStop(cmd *cobra.Command, args []string) {
	dockerInterface := internal.GetDockerInterface(mode)
	if dockerInterface.UseDevInfra {
		fmt.Println("[+] Stopping the development environment")
	} else {
		fmt.Println("[+] Stopping the production environment")
	}

	fmt.Printf("[+] Stopping services with %s...\n", dockerInterface.ComposeFile)
	stopErr := dockerInterface.RunComposeCmd("stop")
	if stopErr != nil {
		log.Fatalf("Error trying to stop services with %s: %v\n", dockerInterface.ComposeFile, stopErr)
	}
}
