package cmd

import (
	"fmt"
	env "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// configTrustOriginCmd represents the configTrustOrigin command
var configTrustOriginCmd = &cobra.Command{
	Use:   "trustorigin <host>",
	Short: "Add a hostname to the trusted origins list",
	Long: `Add a hostname to the trusted origins list. Adding a host to this list will mean that
Ghostwriter will allow requests where the host appears in the "Origin" or "Referer"
headers of requests and does not match the "Host" header.

Use a "*" as a wildcard to trust all subdomains.

Good examples:
	ghostwriter-cli config trustorigin ghostwriter.local
	ghostwriter-cli config trustorigin *.ghostwriter.local
Bad examples:
	ghostwriter-cli config trustorigin *`,
	Args: cobra.ExactArgs(1),
	Run:  configTrustOrigin,
}

func init() {
	configCmd.AddCommand(configTrustOriginCmd)
}

func configTrustOrigin(cmd *cobra.Command, args []string) {
	env.TrustOrigin(args[0])
	fmt.Println("[+] Configuration successfully updated. Bring containers down and up for changes to take effect.")
}
