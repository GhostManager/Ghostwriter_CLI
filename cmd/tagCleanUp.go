package cmd

import (
	"fmt"
	"log"
	"os"

	internal "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
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
	dockerInterface := internal.GetDockerInterface(mode)
	dockerInterface.Env.Save()
	if dockerInterface.UseDevInfra {
		fmt.Println("[+] Executing tag cleanup in the development environment...")
	} else {
		fmt.Println("[+] Executing tag cleanup in the production environment...")
	}

	err := dockerInterface.RunDjangoManageCommand("deduplicate_tags")
	if err != nil {
		log.Fatalf("Could not deduplicate tags: %s\n", err)
	}

	c := internal.AskForConfirmation("[?] Do you want to also remove orphaned tags?")
	if !c {
		os.Exit(0)
	}
	err = dockerInterface.RunDjangoManageCommand("remove_orphaned_tags")
	if err != nil {
		log.Fatalf("Could not remove orphaned tags: %s\n", err)
	}
}
