package cmd

import (
	"github.com/spf13/cobra"
)

// containersCmd represents the containers command
var containersCmd = &cobra.Command{
	Use:   "containers",
	Short: "Manage Ghostwriter containers with subcommands",
	Long: `Manage Ghostwriter containers and services with subcommands. By default, all
subdommands target the production environment.

If you're a developer, use the "--dev" flag to target the development environment.`,
}

func init() {
	rootCmd.AddCommand(containersCmd)
}
