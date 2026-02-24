package cmd

import (
	"fmt"
	"log"

	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// containersBuildCmd represents the build command
var containersBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds the Ghostwriter containers (only needed for updates)",
	Long: `Builds the Ghostwriter containers. Production containers are built by
default. Use the "--mode" argument to build a development environment.

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
	dockerInterface := internal.GetDockerInterface(mode)
	if dockerInterface.UseDevInfra {
		fmt.Println("[+] Starting development environment build")
	} else {
		fmt.Println("[+] Starting production environment build")
	}
	dockerInterface.Env.Save()

	downErr := dockerInterface.Down(nil)
	if downErr != nil {
		log.Fatalf("Error trying to bring down any running containers with %s: %v\n", dockerInterface.ComposeFile, downErr)
	}
	buildErr := dockerInterface.RunComposeCmd("build")
	if buildErr != nil {
		log.Fatalf("Error trying to build with %s: %v\n", dockerInterface.ComposeFile, buildErr)
	}

	upErr := dockerInterface.Up()
	if upErr != nil {
		log.Fatalf("Error trying to bring up environment with %s: %v\n", dockerInterface.ComposeFile, upErr)
	}
	if !skipseed {
		// Must wait for Django to complete any potential db migrations before re-seeding the database
		for {
			if dockerInterface.WaitForDjango() {
				fmt.Println("[+] Re-seeding database in case initial values were added or adjusted...")
				seedErr := dockerInterface.RunComposeCmd("run", "--rm", "django", "/seed_data")
				if seedErr != nil {
					log.Fatalf("Error trying to seed the database: %v\n", seedErr)
				}
				break
			}
		}
	} else {
		fmt.Println("[+] The `--skip-seed` flag was set, so skipped database seeding...")
	}
	fmt.Println("[+] All containers have been built!")
}
