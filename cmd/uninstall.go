package cmd

import (
	"fmt"
	"log"
	"os"

	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
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
	dockerInterface := internal.GetDockerInterface(mode)
	if dockerInterface.UseDevInfra {
		fmt.Println("[+] Starting Ghostwriter development environment removal")
	} else {
		fmt.Println("[+] Starting Ghostwriter production environment removal")
	}
	dockerInterface.Env.Save()

	c := internal.AskForConfirmation("[!] This command removes all containers, images, and volume data for the target environment. Are you sure you want to uninstall?")
	if !c {
		os.Exit(0)
	}
	uninstallErr := dockerInterface.RunComposeCmd("down", "--rmi", "all", "-v", "--remove-orphans")
	if uninstallErr != nil {
		log.Fatalf("Error trying to uninstall with %s: %v\n", dockerInterface.ComposeFile, uninstallErr)
	}
	fmt.Println("[+] Uninstall was successful. You can re-install with `./ghostwriter-cli install`.")
}
