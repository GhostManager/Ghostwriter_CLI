package cmd

import (
	env "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
	"os"
)

// Vars for global flags
var dev bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ghostwriter-cli",
	Short: "A command line interface for managing Ghostwriter.",
	Long: `Ghostwriter CLI is a command line interface for managing the Ghostwriter
application and associated containers and services. Commands are grouped by their use.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Create or parse the Docker ``.env`` file
	env.ParseGhostwriterEnvironmentVariables()

	// Persistent flags defined here are global for the CLI
	rootCmd.PersistentFlags().BoolVar(&dev, "dev", false, `Target the development environment for "install" and "containers" commands.`)
}
