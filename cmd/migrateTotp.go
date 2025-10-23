package cmd

import (
	"bufio"
	"fmt"
	"os"

	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

var migrateTotpCmd = &cobra.Command{
	Use:   "migrate_totp",
	Short: "Migrate TOTP secrets and migration codes from Ghostwriter <=v6 to v6.1+",
	Long: `This command migrates TOTP secrets and migration codes from an installation of Ghostwriter v6.0 or earlier to a Ghostwriter v6.1 or later installation.
It reads the TOTP secrets and migration codes from the database and updates the corresponding user records.
`,
	Run: migrateTotp,
}

func init() {
	rootCmd.AddCommand(migrateTotpCmd)
}

func migrateTotp(cmd *cobra.Command, args []string) {
	var yamlFile string

	docker.EvaluateDockerComposeStatus()
	if dev {
		docker.SetDevMode()
		yamlFile = "local.yml"
	} else {
		docker.SetProductionMode()
		yamlFile = "production.yml"
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Migrating TOTP secrets and migration codes from Ghostwriter <=v6 to v6.1+.\n")
	fmt.Print("Press enter to continue, or Ctrl+C to cancel\n")
	reader.ReadString('\n')

	docker.RunDockerComposeDown(yamlFile, false)
	fmt.Println("[+] migrating TOTP secrets and migration codes")

	docker.RunManagementCmd(yamlFile, "migrate")
	docker.RunManagementCmd(yamlFile, "migrate_totp_device")

	fmt.Println("[+] TOTP secrets and migration codes migration complete")
}
