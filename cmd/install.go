package cmd

import (
	"fmt"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Builds containers and performs first-time setup of Ghostwriter",
	Long: `Builds containers and performs first-time setup of Ghostwriter. A production
environment is installed by default. Use the "--dev" flag to install a development environment.

The command performs the following steps:

* Sets up the default server configuration
* Generates TLS certificates for the server
* Builds the Docker containers
* Creates a default admin user with a randomly generated password

This command only needs to be run once. If you run it again, you will see some errors because
certain actions (e.g., creating the default user) can and should only be done once.`,
	Run: installGhostwriter,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func installGhostwriter(cmd *cobra.Command, args []string) {
	if dev {
		fmt.Println("[+] Starting development environment installation")
		docker.SetDevMode()
		docker.RunDockerComposeInstall("local.yml")
	} else {
		fmt.Println("[+] Starting production environment installation")
		docker.SetProductionMode()
		docker.GenerateCertificatePackage()
		docker.RunDockerComposeInstall("production.yml")
	}

}
