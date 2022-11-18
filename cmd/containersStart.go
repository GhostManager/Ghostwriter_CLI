/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	docker "github.com/GhostManager/Ghostwriter_CLI/cmd/internal"
	"github.com/spf13/cobra"
)

// containersStartCmd represents the start command
var containersStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start all stopped Ghostwriter services",
	Long: `Start all stopped Ghostwriter services. This performs the equivalent
of running the "docker compose start" command.

Production containers are targeted by default. Use the "--dev" flag to
target development containers`,
	Run: containersStart,
}

func init() {
	containersCmd.AddCommand(containersStartCmd)
}

func containersStart(cmd *cobra.Command, args []string) {
	docker.EvaluateDockerComposeStatus()
	if dev {
		fmt.Println("[+] Starting the development environment")
		docker.RunDockerComposeStart("local.yml")
	} else {
		fmt.Println("[+] Starting the production environment")
		docker.RunDockerComposeStart("production.yml")
	}
}
