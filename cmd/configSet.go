package cmd

import (
	"fmt"
	env "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// configSetCmd represents the configSet command
var configSetCmd = &cobra.Command{
	Use:   "set <configuration> <value>",
	Short: "Set the specified configuration value",
	Long: `Set the specified configuration value. Use quotations around the value
if it contains spaces.

For example: ghostwriter-cli config set DATE_FORMAT "d M Y"`,
	Args: cobra.ExactArgs(2),
	Run:  configSet,
}

func init() {
	configCmd.AddCommand(configSetCmd)
}

func configSet(cmd *cobra.Command, args []string) {
	env.SetConfig(args[0], args[1])
	fmt.Println("[+] Configuration successfully updated. Bring containers down and up for changes to take effect.")
}
