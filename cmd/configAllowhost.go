package cmd

import (
	"fmt"
	env "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// configAllowhostCmd represents the configAllowhost command
var configAllowHostCmd = &cobra.Command{
	Use:   "allowhost <host>",
	Short: "Add a hostname or IP address to the allowed hosts list",
	Long: `Add a hostname or IP address to the allowed hosts list. Using a single "*"
as a wildcard to allow all hostnames and IP address will work, but the use of
wildcards in a hostname or address will not work.

Using "*" is NOT recommended! It should only be used for testing purposes.

Good examples:
	ghostwriter-cli config allowhost 192.168.1.100
	ghostwriter-cli config allowhost ghostwriter.local
	ghostwriter-cli config allowhost *
Bad examples:
	ghostwriter-cli config allowhost *.example.com
	ghostwriter-cli config allowhost 192.168.1.*`,
	Args: cobra.ExactArgs(1),
	Run:  configAllowHost,
}

func init() {
	configCmd.AddCommand(configAllowHostCmd)
}

func configAllowHost(cmd *cobra.Command, args []string) {
	env.AllowHost(args[0])
	fmt.Println("[+] Configuration successfully updated. Bring containers down and up for changes to take effect.")
}
