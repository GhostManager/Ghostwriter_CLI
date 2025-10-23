# Ghostwriter_CLI

[![Go](https://img.shields.io/badge/Go-1.24-9cf)](.) [![License](https://img.shields.io/badge/License-BSD3-darkred.svg)](.)

![GitHub Release (Latest by Date)](https://img.shields.io/github/v/release/GhostManager/Ghostwriter_CLI?label=Latest%20Release)
![GitHub Release Date](https://img.shields.io/github/release-date/ghostmanager/Ghostwriter_CLI?label=Release%20Date)

[![CodeFactor](https://img.shields.io/codefactor/grade/github/GhostManager/Ghostwriter_CLI?label=Code%20Quality)](.)

Golang code for the `ghostwriter-cli` binary in [Ghostwriter](https://github.com/GhostManager/Ghostwriter). This binary provides control for various aspects of Ghostwriter's configuration.

Ghostwriter CLI is compatible with Docker Compose v2 and Podman. If using Podman, configure [Docker compatibility mode](https://podman-desktop.io/docs/migrating-from-docker/managing-docker-compatibility).

## Usage

Execute `./ghostwriter-cli help` for usage information (see below). More information about Ghostwriter and how to manage it with `ghostwriter-cli` can be found on the [Ghostwriter Wiki](https://ghostwriter.wiki/).

```plaintext
Ghostwriter CLI is a command line interface for managing the Ghostwriter
application and associated containers and services. Commands are grouped by their use.

Usage:
  ghostwriter-cli [command]

Available Commands:
  backup       Creates a backup of the PostgreSQL database
  completion   Generate the autocompletion script for the specified shell
  config       Display or adjust the configuration
  containers   Manage Ghostwriter containers with subcommands
  down         Shortcut for `containers down`
  gencert      Create a new SSL/TLS certificate and DH param file for the Nginx web server
  healthcheck  Check the health of Ghostwriter's services
  help         Help about any command
  install      Builds containers and performs first-time setup of Ghostwriter
  logs         Fetch logs for Ghostwriter services
  migrate_totp Migrate TOTP secrets and migration codes from Ghostwriter <=v6 to v6.1+
  pg-upgrade   Upgrades the PostgreSQL database
  restore      Restores the specified PostgreSQL database backup
  running      Print a list of running Ghostwriter services
  tagcleanup   Run Django's tag cleanup commands to deduplicate tags and remove orphaned tags
  test         Runs Ghostwriter's unit tests in the development environment
  uninstall    Remove all Ghostwriter containers, images, and volume data
  up           Shortcut for `containers up`
  update       Displays version information for Ghostwriter
  version      Displays Ghostwriter CLI's version information


Flags:
      --dev    Target the development environment for "install" and "containers" commands.
  -h, --help   help for ghostwriter-cli

Use "ghostwriter-cli [command] --help" for more information about a command.
```

## Compilation

The binaries distributed with Ghostwriter and attached to releases are compiled with the following command to set version and build date information:

```bash
go build -ldflags="-s -w -X 'github.com/GhostManager/Ghostwriter_CLI/cmd/config.Version=`git describe --tags --abbrev=0`' -X 'github.com/GhostManager/Ghostwriter_CLI/cmd/config.BuildDate=`date -u '+%d %b %Y'`'" -o ghostwriter-cli main.go
```

The version for rolling releases is set to `rolling`.

Builds pass through `upx` with `upx --brute ghostwriter-cli`. This is simply so that the standard ~8.5MB Golang file is compressed down to a ~2.5MB file for easier inclusion with the Ghostwriter repo.

All releases include the MD5 hash of the packed binary so you can verify the downloaded binary matches the release. Compile and pack the code with the above commands (or use the pre-compiled binary from a release or Ghostwriter) and then calculate the hash with `md5sum` for comparison. Note that hashes will be different if your command sets different version and build date information.
