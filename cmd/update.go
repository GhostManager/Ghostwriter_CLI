package cmd

import (
	"fmt"
	"log"

	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

var updateVersion string

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates and sets up Ghostwriter",
	Long: `Installs and sets up Ghostwriter.

By default, Ghostwriter will download and install the latest version to an application data directory.
Use the '--version' flag to specify a specific version to update to.

If a local '--mode' is specified instead, this command will rebuild the containers and migrate/reseed
the database, but won't actually download a new version (use git fetch+checkout instead).
`,
	Aliases: []string{"upgrade"},
	Run:     updateGhostwriter,
}

func init() {
	updateCmd.PersistentFlags().StringVar(
		&updateVersion,
		"version",
		"",
		"Version to install. Defaults to the latest tagged release. Ignored for --mode=local-*. NOTE: downgrading is not supported.",
	)
	rootCmd.AddCommand(updateCmd)
}

func updateGhostwriter(cmd *cobra.Command, args []string) {
	var err error

	if mode == docker.ModeProd {
		// Fetch and write docker-compose.yml file
		err = fetchAndWriteComposeFile(mode, updateVersion)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}

	// Get interface
	dockerInterface := docker.GetDockerInterface(mode)
	dockerInterface.Env.Save()
	if dockerInterface.UseDevInfra {
		fmt.Println("[+] Starting development environment update")
	} else {
		fmt.Println("[+] Starting production environment update")
		docker.PrepareSettingsDirectory(dockerInterface.Dir)
	}

	fmt.Println("[+] Tearing down containers...")
	err = dockerInterface.Down(&docker.DownOptions{
		RemoveOrphans: true,
	})
	if err != nil {
		log.Fatalf("Could not tear down containers: %v", err)
	}

	err = updateContainers(*dockerInterface)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	fmt.Println("[+] Starting containers...")
	err = dockerInterface.Up()
	if err != nil {
		log.Fatalf("Error bringing containers up: %s\n", err)
	}

	fmt.Println("[+] Ghostwriter is ready to go!")
}
