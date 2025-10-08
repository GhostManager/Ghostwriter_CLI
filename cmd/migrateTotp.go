package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"

	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

var migrateTotpCmd = &cobra.Command{
	Use:   "migrate_totp",
	Short: "Migrate TOTP secrets and migration codes from Ghostwriter v1 to v2",
	Long: `This command migrates TOTP secrets and migration codes from a Ghostwriter v1 installation to a Ghostwriter v2 installation.
It reads the TOTP secrets and migration codes from the v1 database and updates the corresponding user records in the v2 database.
`,
	Run: migrateTotp,
}


func init() {
	rootCmd.AddCommand(migrateTotpCmd)
}

func migrateTotp(cmd *cobra.Command, args []string){
	docker.EvaluateDockerComposeStatus()
	var yamlFile string
	if dev {
		docker.SetDevMode()
		yamlFile = "local.yml"
	} else {
		docker.SetProductionMode()
		yamlFile = "production.yml"
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Migrating TOTP secrets and migration codes from Ghostwriter v1 to v2.\n")
	fmt.Print("Press enter to continue, or Ctrl+C to cancel\n")
	reader.ReadString('\n')

	docker.RunDockerComposeDown(yamlFile, false)
	fmt.Println("[+] migrating TOTP secrets and migration codes")


	err := docker.RunCmd("docker", []string{"-f", yamlFile, "run", "--rm", "django", "python", "manage.py", "makemigrations"})
	if err != nil {
		log.Fatalf("Error occurred while running makemigrations: %v", err)
	}
	err = docker.RunCmd("docker", []string{"-f", yamlFile, "run", "--rm", "django", "python", "manage.py", "migrate"})
	if err != nil {
		log.Fatalf("Error occurred while running migrate: %v", err)
	}
	err = docker.RunCmd("docker", []string{"-f", yamlFile, "run", "--rm", "django", "python", "manage.py", "migrate_totp_device"})
	if err != nil {
		log.Fatalf("Error occurred while running migrate_totp_device: %v", err)
	}

	fmt.Println("[+] TOTP secrets and migration codes migration complete")
	fmt.Println("Starting services...")
	docker.RunDockerComposeUp(yamlFile)
	fmt.Println("Services started.")
}