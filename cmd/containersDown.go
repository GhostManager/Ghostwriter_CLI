package cmd

import (
	"fmt"
	"log"

	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

var volumes bool

// containersDownCmd represents the down command
var containersDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Bring down all Ghostwriter services and remove the containers",
	Long: `Bring down all Ghostwriter services and remove the containers. This
performs the equivalent of running the "docker compose down" command.

Production containers are targeted by default. Use the "--mode" argument to
target development containers`,
	Run: containersDown,
}

func init() {
	containersCmd.AddCommand(containersDownCmd)

	containersDownCmd.PersistentFlags().BoolVar(&volumes, "volumes", false, "Delete data volumes when containers come down")
}

func containersDown(cmd *cobra.Command, args []string) {
	dockerInterface := internal.GetDockerInterface(mode)
	if dockerInterface.UseDevInfra {
		fmt.Println("[+] Bringing down the development environment")
	} else {
		fmt.Println("[+] Bringing down the production environment")
	}
	err := dockerInterface.Down(&internal.DownOptions{
		Volumes: volumes,
	})
	if err != nil {
		log.Fatalf("Error trying to bring down the containers with %s: %v\n", dockerInterface.ComposeFile, err)
	}
}
