package cmd

import (
	"fmt"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove all Ghostwriter containers, images, and volume data",
	Long: `Remove all Ghostwriter containers, images, and volume data.

The command performs the following steps:

* Brings down running containers
* Deletes the stopped containers
* Deletes the container images
* Deletes all Ghostwriter volumes and data

This command is irreversible and should only be run if you are looking to remove Ghostwriter from the system or wanting
a fresh start for the target environment.`,
	Run: uninstallGhostwriter,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func uninstallGhostwriter(cmd *cobra.Command, args []string) {
	docker.EvaluateDockerComposeStatus()
	if dev {
		fmt.Println("[+] Starting Ghostwriter development environment removal")
		docker.SetDevMode()
		docker.RunDockerComposeUninstall("local.yml")
	} else {
		fmt.Println("[+] Starting Ghostwriter production environment removal")
		docker.SetProductionMode()
		docker.RunDockerComposeUninstall("production.yml")
	}

}
