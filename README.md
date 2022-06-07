# Ghostwriter_CLI

[![Go](https://img.shields.io/badge/Go-1.18-9cf)](.) [![License](https://img.shields.io/badge/License-BSD3-darkred.svg)](.)

![GitHub Release (Latest by Date)](https://img.shields.io/github/v/release/GhostManager/Ghostwriter_CLI?label=Latest%20Release)
![GitHub Release Date](https://img.shields.io/github/release-date/ghostmanager/Ghostwriter_CLI?label=Release%20Date)

[![CodeFactor](https://img.shields.io/codefactor/grade/github/GhostManager/Ghostwriter_CLI?label=Code%20Quality)](.)

Golang code for the `ghostwriter-cli` binary in [Ghostwriter](https://github.com/GhostManager/Ghostwriter). This binary provides control for various aspects of Ghostwriter's configuration.

## Usage

Execute `./ghostwriter-cli help` for usage information (see below). More information about Ghostwriter and how to manage it with `ghostwriter-cli` can be found on the [Ghostwriter Wiki](https://ghostwriter.wiki/).

```
Ghostwriter-CLI ( v0.1.0, 7 June 2022 ):
********************************************************************
*** source code: https://github.com/GhostManager/Ghostwriter_CLI ***
********************************************************************
  help
    Displays this help information
  install {dev|production}
    Builds containers and performs first-time setup of Ghostwriter
  build {dev|production}
    Builds the containers for the given environment (only necessary for upgrades)
  restart {dev|production}
    Restarts all Ghostwriter services in the given environment
  up {dev|production}
    Bring up all Ghostwriter services in the given environment
  down {dev|production}
    Bring down all Ghostwriter services in the given environment
  config
    ** No parameters will dump the entire config **
    get [varname ...]
    set <var name> <var value>
    allowhost <var hostname/address>
    disallowhost <var hostname/address>
  logs <container name>
    Displays logs for the given container
    Options: ghostwriter_{django|nginx|postgres|redis|graphql|queue}
  running
    Print a list of running Ghostwriter services
  update
    Displays version information for the local Ghostwriter installation and the latest stable release on GitHub
  test
    Runs Ghostwriter's unit tests in the development environment
    Requires to `ghostwriter_cli install dev` to have been run first
  version
    Displays the version information at the top of this message
```

## Compilation

The binaries distributed with Ghostwriter and attached to releases are compiled with `go build -ldflags="-s -w" -o ghostwriter-cli ghostwriter-cli.go` and then passed through `upx` with `upx --brute ghostwriter-cli`. This is simply so that the standard ~8.5MB Golang file is compressed down to a ~2.5MB file for easier inclusion with the Ghostwriter repo.

All releases include the MD5 hash of the packed binary so you can verify the downloaded binary matches the release. Compile and pack the code with the above commands (or use the pre-compiled binary from a release or Ghostwriter) and then calculate the hash with `md5sum` for comparison.
