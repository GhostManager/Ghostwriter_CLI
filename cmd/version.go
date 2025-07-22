package cmd

import (
	"fmt"
	"github.com/GhostManager/Ghostwriter_CLI/cmd/config"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays Ghostwriter CLI's version information",
	Long:  "Displays Ghostwriter CLI's version information.",
	Run:   displayVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func displayVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("Ghostwriter-CLI %s ( %s )\n", config.Version, config.BuildDate)
}
