package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"

	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
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
	dockerInterface := internal.GetDockerInterface(mode)
	dockerInterface.Env.Save()
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Migrating TOTP secrets and migration codes from Ghostwriter <=v6 to v6.1+.\n")
	fmt.Print("Press enter to continue, or Ctrl+C to cancel\n")
	reader.ReadString('\n')

	err := dockerInterface.Down(nil)
	if err != nil {
		log.Fatalf("Error trying to bring down the containers with %s: %v\n", dockerInterface.ComposeFile, err)
	}

	fmt.Println("[+] migrating TOTP secrets and migration codes")
	err = dockerInterface.RunDjangoManageCommand("migrate")
	if err != nil {
		log.Fatalf("Error trying to migrate the database with %s: %v\n", dockerInterface.ComposeFile, err)
	}
	err = dockerInterface.RunDjangoManageCommand("migrate_totp_device")
	if err != nil {
		log.Fatalf("Error trying to migrate the TOTP devices with %s: %v\n", dockerInterface.ComposeFile, err)
	}

	fmt.Println("[+] TOTP secrets and migration codes migration complete")
}
