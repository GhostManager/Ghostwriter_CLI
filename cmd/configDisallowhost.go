package cmd

import (
	"fmt"
	"log"

	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// configDisallowhostCmd represents the configDisallowhost command
var configDisallowHostCmd = &cobra.Command{
	Use:   "disallowhost <host>",
	Short: "Remove a hostname or IP address to the allowed hosts list",
	Long:  "Remove a hostname or IP address to the allowed hosts list.",
	Args:  cobra.ExactArgs(1),
	Run:   configDisallowHost,
}

func init() {
	configCmd.AddCommand(configDisallowHostCmd)
}

func configDisallowHost(cmd *cobra.Command, args []string) {
	env, err := internal.ReadEnv(internal.GetDockerDirFromMode(mode))
	if err != nil {
		log.Fatalf("Could not read environment file: %s\n", err)
	}
	env.RemoveHost("django_allowed_hosts", args[0])
	env.Save()
	fmt.Println("[+] Configuration successfully updated. Bring containers down and up for changes to take effect.")
}
