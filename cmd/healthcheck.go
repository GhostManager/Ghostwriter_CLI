package cmd

import (
	"fmt"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	utils "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

// healthcheckCmd represents the healthcheck command
var healthcheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "Check the health of Ghostwriter's services",
	Long: `Check the health of Ghostwriter's services.

This command validates all containers are running and passing
their respective health checks`,
	Run: runHealthcheck,
}

func init() {
	rootCmd.AddCommand(healthcheckCmd)
}

func runHealthcheck(cmd *cobra.Command, args []string) {
	docker.EvaluateDockerComposeStatus()
	// initialize tabwriter
	writer := new(tabwriter.Writer)
	// Set minwidth, tabwidth, padding, padchar, and flags
	writer.Init(os.Stdout, 8, 8, 1, '\t', 0)

	defer writer.Flush()

	fmt.Println("[+] Checking Ghostwriter containers and their respective health checks...")

	containerIssues, dockerErr := docker.CheckDockerHealth(dev)

	if dockerErr != nil {
		fmt.Printf("[!] Failed to get container information from Docker: %s\n", dockerErr)
	} else {
		if len(containerIssues) > 0 {
			fmt.Printf("[*] Identified %d issues with one or more containers:\n\n", len(containerIssues))

			fmt.Fprintf(writer, "\n %s\t%s\t%s", "Type", "Container", "Message")
			fmt.Fprintf(writer, "\n %s\t%s\t%s", "––––––––––––", "––––––––––––", "––––––––––––")

			for _, issue := range containerIssues {
				fmt.Fprintf(writer, "\n %s\t%s\t%s", issue.Type, issue.Service, issue.Message)
			}
		} else {
			fmt.Println("[*] Identified zero container issues, now testing services...")
			serviceIssues, svcErr := utils.CheckGhostwriterHealth(dev)
			if svcErr != nil {
				fmt.Printf("[!] Failed to get health status from Ghostwriter's /status/ endpoint: %s\n", svcErr)
			} else {
				if len(serviceIssues) > 0 {
					fmt.Printf("[*] Identified %d issues with one or more services:\n\n", len(serviceIssues))

					fmt.Fprintf(writer, "\n %s\t%s\t%s", "Type", "Service", "Message")
					fmt.Fprintf(writer, "\n %s\t%s\t%s", "––––––––––––", "––––––––––––", "––––––––––––")

					for _, issue := range serviceIssues {
						fmt.Fprintf(writer, "\n %s\t%s\t%s", issue.Type, issue.Service, issue.Message)
					}
				} else {
					fmt.Println("[*] Identified zero issues with core services")
				}
			}
		}
	}
}
