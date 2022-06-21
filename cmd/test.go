package cmd

import (
	"fmt"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs Ghostwriter's unit tests in the development environment",
	Long: `Runs Ghostwriter's unit tests in the development environment.

Requires to "install --dev" to have been run first.`,
	Run: runUnitTests,
}

func init() {
	rootCmd.AddCommand(testCmd)
}

func runUnitTests(cmd *cobra.Command, args []string) {
	fmt.Println("[+] Running Ghostwriter's unit and integration tests...")
	docker.RunGhostwriterTests()
}
