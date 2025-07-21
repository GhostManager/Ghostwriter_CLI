package cmd

import (
	"fmt"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
	"os"
)

// containersUpCmd represents the up command
var tagCleanUpCmd = &cobra.Command{
	Use:   "tagcleanup",
	Short: "Run Django's tag cleanup commands to deduplicate tags and remove orphaned tags",
	Long: `Run Django's tag cleanup commands to deduplicate tags and remove orphaned tags, including:

* remove_orphaned_tags
* deduplicate_tags

When deduplicating tags, the tag with the oldest primary key value (the first created) will be kept.

Note: These commands are only available with Ghostwriter v6 or later.`,
	Run: tagCleanUp,
}

func init() {
	rootCmd.AddCommand(tagCleanUpCmd)
}

func tagCleanUp(cmd *cobra.Command, args []string) {
	docker.EvaluateDockerComposeStatus()
	if dev {
		fmt.Println("[+] Executing tag cleanup in the development environment...")
		docker.SetDevMode()
		docker.RunManagementCmd("local.yml", "deduplicate_tags")
		c := docker.AskForConfirmation("[?] Do you want to also remove orphaned tags?")
		if !c {
			os.Exit(0)
		}
		docker.RunManagementCmd("local.yml", "remove_orphaned_tags")
	} else {
		fmt.Println("[+] Executing tag cleanup in the production environment...")
		docker.SetProductionMode()
		docker.RunManagementCmd("local.yml", "deduplicate_tags")
		c := docker.AskForConfirmation("[?] Do you want to also remove orphaned tags?")
		if !c {
			os.Exit(0)
		}
		docker.RunManagementCmd("production.yml", "remove_orphaned_tags")
	}
}
