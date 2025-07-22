package cmd

import (
	"fmt"
	"github.com/GhostManager/Ghostwriter_CLI/cmd/config"
	utils "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays Ghostwriter CLI's version information",
	Long: `Displays Ghostwriter CLI's version information. The local version information comes from the current binary.
The latest release information is pulled from GitHub's API`,
	RunE: compareCliVersions,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func compareCliVersions(cmd *cobra.Command, args []string) error {
	// initialize tabwriter
	writer := new(tabwriter.Writer)
	// Set minwidth, tabwidth, padding, padchar, and flags
	writer.Init(os.Stdout, 8, 8, 1, '\t', 0)

	defer writer.Flush()

	fmt.Println("[+] Fetching latest version information:")

	if len(config.BuildDate) == 0 {
		fmt.Fprintf(writer, "\nLocal Version\tGhostwriter CLI %s", config.Version)
	} else {
		fmt.Fprintf(writer, "\nLocal Version\tGhostwriter CLI %s (%s)", config.Version, config.BuildDate)
	}

	remoteVersion, htmlUrl, remoteErr := utils.GetRemoteVersion("GhostManager", "Ghostwriter_CLI")
	if remoteErr != nil {
		return remoteErr
	}

	fmt.Fprintf(writer, "\nLatest Release\t%s\n", remoteVersion)
	fmt.Fprintf(writer, "Latest Download URL\t%s\n", htmlUrl)

	return nil
}
