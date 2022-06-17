package config

// Constants and variables used by the Ghostwriter CLI

var (
	// Ghostwriter CLI version
	// This gets populated at build time with the following flags:
	// -ldflags "-X 'github.com/GhostManager/Ghostwriter_CLI/cmd/config.Version=`git describe --tags --abbrev=0`'
	//     -X 'github.com/GhostManager/Ghostwriter_CLI/cmd/config.BuildDate=`date -u '+%d %b %Y'`'"
	Version     string = "v0.1.0"
	BuildDate   string
	Name        string = "Ghostwriter CLI"
	DisplayName string = "Ghostwriter CLI"
	Description string = "A command line interface for Ghostwriter"
)
