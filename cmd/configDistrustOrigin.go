package cmd

import (
	"fmt"
	env "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// configDistrustOriginCmd represents the configDistrustOrigin command
var configDistrustOriginCmd = &cobra.Command{
	Use:   "distrustorigin <host>",
	Short: "Remove a hostname from the trusted origins list",
	Long: `Remove a hostname from the trusted origins list. Removing a host from this list will
mean that Ghostwriter will block requests where the host appears in the "Origin" or
"Referer" headers of requests and does not match the "Host" header.`,
	Args: cobra.ExactArgs(1),
	Run:  configDistrustOrigin,
}

func init() {
	configCmd.AddCommand(configDistrustOriginCmd)
}

func configDistrustOrigin(cmd *cobra.Command, args []string) {
	env.DistrustOrigin(args[0])
	fmt.Println("[+] Configuration successfully updated. Bring containers down and up for changes to take effect.")
}
