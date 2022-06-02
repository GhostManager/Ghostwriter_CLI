# Ghostwriter_CLI

[![Go](https://img.shields.io/badge/Go-1.18-9cf)](.) [![License](https://img.shields.io/badge/License-BSD3-darkred.svg)](.)

![GitHub Release (Latest by Date)](https://img.shields.io/github/v/release/GhostManager/Ghostwriter-CLI?label=Latest%20Release)
![GitHub Release Date](https://img.shields.io/github/release-date/ghostmanager/Ghostwriter-CLI?label=Release%20Date)

[![CodeFactor](https://img.shields.io/codefactor/grade/github/GhostManager/Ghostwriter-CLI?label=Code%20Quality)](.)

Golang code for the `ghostwriter-cli` binary in [Ghostwriter](https://github.com/GhostManager/Ghostwriter). This binary provides control for various aspects of Ghostwriter's configuration.

## Usage

Execute `./ghostwriter-cli help` for usage information (see below). More information about Ghostwriter and how to manage it with `ghostwriter-cli` can be found on the [Ghostwriter Wiki](https://ghostwriter.wiki/).

```
Ghostwriter-CLI ( v0.0.2, 2 June 2022 ):
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
  logs <container name>
    Displays logs for the given container
    Options: ghostwriter_{django|nginx|postgres|redis|graphql|queue}
  update
    Displays version information for the local Ghostwriter installation and the latest stable release on GitHub
  test
    Runs Ghostwriter's unit tests in the development environment
    Requires to `ghostwriter_cli install dev` to have been run first
  version
    Displays the version information at the top of this message
```

## Compilation

The binary distributed with Ghostwriter and attached to releases is compiled with `go build -ldflags="-s -w" -o ghostwriter-cli ghostwriter-cli.go` and then passed through `upx` with `upx --brute ghostwriter-cli`. This is simply so that the standard ~8.5MB Golang file is compressed down to a ~2.5MB file for easier inclusion with the Ghostwriter repo.

All releases include the SHA512 hash of the packed binary so you can verify the binary matches the release. Compile and pack the code with the above commands (or use the pre-compiled binary from a release or Ghostwriter) and then run the following command to get the hash for comparison: `shasum -a 512 ghostwriter-cli`

## Release Checklist

To create a new release of *Ghostwriter_CLI*, follow the steps below:

1. Build the binary: `go build -ldflags="-s -w" -o ghostwriter-cli ghostwriter-cli.go`
2. Pack the binary with `upx` to reduce the size: `upx --brute ghostwriter-cli`
3. Record the SHA512 hash of the packed binary: `shasum -a 512 ghostwriter-cli`
4. Create a new release
