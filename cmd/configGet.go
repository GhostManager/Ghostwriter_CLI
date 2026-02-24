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

// configGetCmd represents the configGet command
var configGetCmd = &cobra.Command{
	Use:   "get <configuration> <configuration> ...",
	Short: "Get the specified configuration values",
	Long: `Get the specified configuration values. You can provide one value or
a list of values separated by spaces.

For example: ghostwriter-cli config get ADMIN_PASSWORD POSTGRES_PASSWORD`,
	Run: configGet,
}

func init() {
	configCmd.AddCommand(configGetCmd)
}

func configGet(cmd *cobra.Command, args []string) {
	// initialize tabwriter
	writer := new(tabwriter.Writer)
	// Set minwidth, tabwidth, padding, padchar, and flags
	writer.Init(os.Stdout, 8, 8, 1, '\t', 0)

	defer writer.Flush()

	env, err := internal.ReadEnv(internal.GetDockerDirFromMode(mode))
	if err != nil {
		log.Fatalf("Could not read environment file: %s\n", err)
	}

	fmt.Println("[+] Getting configuration values:")
	fmt.Fprintf(writer, "\n %s\t%s", "Setting", "Value")
	fmt.Fprintf(writer, "\n %s\t%s", "–––––––", "–––––––")

	for _, arg := range args {
		val := env.Get(arg)
		if val == "" {
			val = "–"
		}
		fmt.Fprintf(writer, "\n %s\t%s", strings.ToUpper(arg), val)
	}
	fmt.Fprintln(writer, "")
}
