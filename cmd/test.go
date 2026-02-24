package cmd

import (
	"fmt"

	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs Ghostwriter's unit tests in the development environment",
	Long: `Runs Ghostwriter's unit tests in the development environment.

Requires to "install --mode=local-dev" to have been run first.`,
	RunE: runUnitTests,
}

func init() {
	rootCmd.AddCommand(testCmd)
}

func runUnitTests(cmd *cobra.Command, args []string) error {
	dockerInterface := internal.GetDockerInterface(mode)
	dockerInterface.Env.Save()
	fmt.Println("[+] Running Ghostwriter's unit and integration tests...")

	// Save the current env values we're about to change
	currentActionSecret := dockerInterface.Env.Get("HASURA_GRAPHQL_ACTION_SECRET")
	currentSettingsModule := dockerInterface.Env.Get("DJANGO_SETTINGS_MODULE")

	// Defer restoration to ensure it happens even if tests fail
	defer func() {
		dockerInterface.Env.Set("HASURA_GRAPHQL_ACTION_SECRET", currentActionSecret)
		dockerInterface.Env.Set("DJANGO_SETTINGS_MODULE", currentSettingsModule)
		dockerInterface.Env.Save()
	}()

	// Change env values for the test conditions
	dockerInterface.Env.Set("HASURA_GRAPHQL_ACTION_SECRET", "changeme")
	dockerInterface.Env.Set("DJANGO_SETTINGS_MODULE", "config.settings.local")
	dockerInterface.Env.Save()

	// Run the unit tests
	testErr := dockerInterface.RunDjangoManageCommand("test")
	if testErr != nil {
		return fmt.Errorf("error trying to run Ghostwriter's tests: %w", testErr)
	}

	return nil
}
