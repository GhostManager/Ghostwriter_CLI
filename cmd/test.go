package cmd

import (
	"fmt"
	"log"

	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs Ghostwriter's unit tests in the development environment",
	Long: `Runs Ghostwriter's unit tests in the development environment.

Requires to "install --mode=local-dev" to have been run first.`,
	Run: runUnitTests,
}

func init() {
	rootCmd.AddCommand(testCmd)
}

func runUnitTests(cmd *cobra.Command, args []string) {
	dockerInterface := internal.GetDockerInterface(mode)
	dockerInterface.Env.Save()
	fmt.Println("[+] Running Ghostwriter's unit and integration tests...")

	// Save the current env values we're about to change
	currentActionSecret := dockerInterface.Env.Get("HASURA_GRAPHQL_ACTION_SECRET")
	currentSettingsModule := dockerInterface.Env.Get("DJANGO_SETTINGS_MODULE")

	// Change env values for the test conditions
	dockerInterface.Env.Set("HASURA_GRAPHQL_ACTION_SECRET", "changeme")
	dockerInterface.Env.Set("DJANGO_SETTINGS_MODULE", "config.settings.local")
	dockerInterface.Env.Save()

	// Run the unit tests
	testErr := dockerInterface.RunDjangoManageCommand("test")
	if testErr != nil {
		log.Fatalf("Error trying to run Ghostwriter's tests: %v\n", testErr)
	}

	// Reset the changed env values
	dockerInterface.Env.Set("HASURA_GRAPHQL_ACTION_SECRET", currentActionSecret)
	dockerInterface.Env.Set("DJANGO_SETTINGS_MODULE", currentSettingsModule)
	dockerInterface.Env.Save()
}
