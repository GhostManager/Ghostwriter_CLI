package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/GhostManager/Ghostwriter_CLI/cmd/config"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	utils "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
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

	dockerInterface := docker.GetDockerInterface(mode)
	dockerCurrentVersion, err := dockerInterface.GetVersion()
	if err != nil {
		return err
	}

	gwcliLatestVersion, htmlUrl, err := utils.GetRemoteVersion("GhostManager", "Ghostwriter_CLI")
	if err != nil {
		return err
	}

	dockerLatestVersion, _, err := utils.GetRemoteVersion("GhostManager", "Ghostwriter")
	if err != nil {
		return err
	}

	fmt.Println()

	fmt.Fprintf(writer, "Ghostwriter CLI\n")
	if len(config.BuildDate) == 0 {
		fmt.Fprintf(writer, "Local Version\t%s\n", config.Version)
	} else {
		fmt.Fprintf(writer, "Local Version\t%s (%s)\n", config.Version, config.BuildDate)
	}
	fmt.Fprintf(writer, "Latest Release\t%s\n", gwcliLatestVersion)
	fmt.Fprintf(writer, "\n")

	fmt.Fprintf(writer, "Ghostwriter\n")
	fmt.Fprintf(writer, "Local Version\t%s\n", dockerCurrentVersion)
	fmt.Fprintf(writer, "Latest Release\t%s\n", dockerLatestVersion)
	fmt.Fprintf(writer, "\n")

	if gwcliLatestVersion != config.Version {
		fmt.Fprintf(writer, "Download the latest version of Ghostwriter CLI at:\t%s\n", htmlUrl)
	}
	if dockerLatestVersion != dockerCurrentVersion {
		fmt.Fprintf(writer, "Install the latest version of Ghostwriter using the `update` subcommand\n")
	}

	if dockerCurrentVersion == "latest" {
		fmt.Println("[!] Using the `latest` tag is not recommended - pulling containers will not apply necessary changes to the docker-compose.yml file")
	}

	writer.Flush()

	return nil
}
