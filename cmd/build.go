package cmd

import (
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Shortcut for `containers build`",
	Run: func(cmd *cobra.Command, args []string) {
		containersBuildCmd.Run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVar(
		&skipseed,
		"skip-seed",
		false,
		`Skip (re-)seeding the database. This is useful when upgrading an existing installation and you know there are no new or adjusted values.`,
	)
}
