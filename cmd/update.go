package cmd

import (
	"fmt"
	utils "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Displays version information for Ghostwriter.",
	Long: `Displays version information for Ghostwriter. The local version
information comes from the local "VERSION" file. The latest release
information is pulled from GitHub's API.`,
	RunE: compareVersions,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func compareVersions(cmd *cobra.Command, args []string) error {
	// initialize tabwriter
	writer := new(tabwriter.Writer)
	// Set minwidth, tabwidth, padding, padchar, and flags
	writer.Init(os.Stdout, 8, 8, 1, '\t', 0)

	defer writer.Flush()

	fmt.Println("[+] Fetching latest version information:")

	localVersion, localErr := utils.GetLocalGhostwriterVersion()
	if localErr != nil {
		return localErr
	}

	remoteVersion, remoteErr := utils.GetRemoteGhostwriterVersion()
	if remoteErr != nil {
		return remoteErr
	}
	fmt.Fprintf(writer, "\nLocal Ghostwriter Version\t%s", localVersion)
	fmt.Fprintf(writer, "Latest Stable Release\t%s", remoteVersion)

	return nil
}
