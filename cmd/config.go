package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display or adjust the configuration",
	Long: `Run this command to display the configuration. Use subcommands to
adjust the configuration or retrieve individual values.`,
	Run: configDisplay,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func configDisplay(cmd *cobra.Command, args []string) {
	env, err := internal.ReadEnv(internal.GetDockerDirFromMode(mode))
	if err != nil {
		log.Fatalf("Could not read environment file: %s\n", err)
	}

	// initialize tabwriter
	writer := new(tabwriter.Writer)
	// Set minwidth, tabwidth, padding, padchar, and flags
	writer.Init(os.Stdout, 8, 8, 1, '\t', 0)

	defer writer.Flush()

	fmt.Println("[+] Current configuration and available variables:")
	fmt.Fprintf(writer, "\n %s\t%s", "Setting", "Value")
	fmt.Fprintf(writer, "\n %s\t%s", "–––––––", "–––––––")

	configuration := env.GetAll()
	for _, config := range configuration {
		if config.Val == "" {
			config.Val = "–"
		}
		fmt.Fprintf(writer, "\n %s\t%s", strings.ToUpper(config.Key), config.Val)
	}
	fmt.Fprintln(writer, "")
}
