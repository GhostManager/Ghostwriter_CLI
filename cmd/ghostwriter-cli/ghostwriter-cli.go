package main

import (
	"fmt"
	"github.com/GhostManager/Ghostwriter_CLI/internal"
	"log"
	"os"
)

var ghostwriterCliVersion = "0.0.1"
var ghostwriterCliBuildDate = "10 May 2022"

func main() {
	// Display help if no arguments are passed
	if len(os.Args) <= 1 {
		internal.DisplayHelp(ghostwriterCliVersion, ghostwriterCliBuildDate)
		os.Exit(0)
	}

	// Create or parse the the Docker ``.env`` file
	internal.ParseGhostwriterEnvironmentVariables()

	switch os.Args[1] {

	// Display help
	case "help":
		internal.DisplayHelp(ghostwriterCliVersion, ghostwriterCliBuildDate)

	// Install Ghostwriter
	case "install":
		if len(os.Args) <= 2 {
			log.Fatalf(
				"missing subcommand for %s; should be 'dev' or 'production'",
				os.Args[1],
			)
		}
		if os.Args[2] == "dev" {
			fmt.Println("[+] Starting development environment installation")
			internal.SetDevMode()
			internal.RunDockerComposeInstall("local.yml")
		} else if os.Args[2] == "production" {
			internal.SetProductionMode()
			fmt.Println("[+] Starting production environment installation")
			internal.GenerateCertificatePackage()
			internal.RunDockerComposeInstall("production.yml")
		} else {
			log.Fatalf("unknown install type; should be 'dev' or 'production'")
		}

	// Rebuild the Ghostwriter containers for upgrades or code changes
	case "build":
		if len(os.Args) <= 2 {
			log.Fatalf(
				"missing subcommand for %s; should be 'dev' or 'production'",
				os.Args[1],
			)
		}

		if os.Args[2] == "dev" {
			fmt.Println("[+] Starting development environment build")
			internal.SetDevMode()
			internal.RunDockerComposeUpgrade("local.yml")
		} else if os.Args[2] == "production" {
			fmt.Println("[+] Starting production environment build")
			internal.SetProductionMode()
			internal.RunDockerComposeUpgrade("production.yml")
		} else {
			log.Fatalf("unknown install type; should be 'dev' or 'production'")
		}

	// Restart the Ghostwriter containers
	case "restart":
		if len(os.Args) <= 2 {
			log.Fatalf(
				"missing subcommand for %s; should be 'dev' or 'production'",
				os.Args[1],
			)
		}

		if os.Args[2] == "dev" {
			fmt.Println("[+] Restarting development environment")
			internal.RunDockerComposeRestart("local.yml")
		} else if os.Args[2] == "production" {
			fmt.Println("[+] Restarting production environment")
			internal.RunDockerComposeRestart("local.yml")
		} else {
			log.Fatalf("unknown environment type; should be 'dev' or 'production'")
		}

	// Bring up all Ghostwriter containers
	case "up":
		if len(os.Args) <= 2 {
			log.Fatalf(
				"missing subcommand for %s; should be 'dev' or 'production'",
				os.Args[1],
			)
		}

		if os.Args[2] == "dev" {
			fmt.Println("[+] Bringing up development environment")
			internal.RunDockerComposeUp("local.yml")
		} else if os.Args[2] == "production" {
			fmt.Println("[+] Bringing up production environment")
			internal.RunDockerComposeUp("production.yml")
		} else {
			log.Fatalf("unknown environment type; should be 'dev' or 'production'")
		}

	// Bring down all Ghostwriter containers
	case "down":
		if len(os.Args) <= 2 {
			log.Fatalf(
				"missing subcommand for %s; should be 'dev' or 'production'",
				os.Args[1],
			)
		}

		if os.Args[2] == "dev" {
			fmt.Println("[+] Stopping development environment")
			internal.RunDockerComposeDown("local.yml")
		} else if os.Args[2] == "production" {
			fmt.Println("[+] Stopping production environment")
			internal.RunDockerComposeDown("production.yml")
		} else {
			log.Fatalf("unknown environment type; should be 'dev' or 'production'")
		}

	// Print the current config or process ``get``and ``set`` commands
	case "config":
		internal.Env(os.Args[2:])

	// Fetch and print logs for the given container
	case "logs":
		logs := internal.FetchLogs(os.Args[2])
		for _, entry := range logs {
			fmt.Print(entry)
		}

	// Get list of running Ghostwriter containers
	case "running":
		internal.GetRunning()

	// Check for Ghostwriter updates
	case "update":
		internal.GetLocalGhostwriterVersion()
		internal.GetRemoteGhostwriterVersion()

	// Run Ghostwriter's unit tests
	case "test":
		internal.RunGhostwriterTests()

	// Print the Ghostwriter_CLI version info
	case "version":
		fmt.Printf("Ghostwriter-CLI ( v%s, %s )\n", ghostwriterCliVersion, ghostwriterCliBuildDate)

	// Display help
	default:
		internal.DisplayHelp(ghostwriterCliVersion, ghostwriterCliBuildDate)
	}
}
