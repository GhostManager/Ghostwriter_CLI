package cmd

import (
	"fmt"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// containersBuildCmd represents the build command
var containersBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds the Ghostwriter containers (only needed for updates)",
	Long: `Builds the Ghostwriter containers. Production containers are built by
default. Use the "--dev" flag to build a development environment.

Note: Build will stop a container if it is already running. You will need to run
the "up" command to start the containers after the build.

Running this command is only necessary when upgrading an existing Ghostwriter installation.`,
	Run: buildContainers,
}

var skipseed bool

func init() {
	containersCmd.AddCommand(containersBuildCmd)
	containersBuildCmd.Flags().BoolVar(
		&skipseed,
		"skip-seed",
		false,
		`Skip (re-)seeding the database. This is useful when upgrading an existing and you know there are no new or adjusted values.`,
	)
}

func buildContainers(cmd *cobra.Command, args []string) {
	docker.EvaluateDockerComposeStatus()
	if dev {
		fmt.Println("[+] Starting development environment build")
		docker.SetDevMode()
		docker.RunDockerComposeUpgrade("local.yml", skipseed)
	} else {
		fmt.Println("[+] Starting production environment build")
		docker.SetProductionMode()
		docker.RunDockerComposeUpgrade("production.yml", skipseed)
	}

}
