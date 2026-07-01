package internal

import (
	"fmt"
	"strings"
)

// SeedDatabase runs Ghostwriter's /seed_data helper with optional arguments.
//
// Some older Ghostwriter images and local source checkouts do not support newer
// /seed_data flags, and the CLI cannot reliably infer support from VERSION or
// the network. When an optional argument is rejected by the command-line parser,
// retry once without optional arguments so updates can proceed against older
// builds. Other seed failures still return as hard errors.
func SeedDatabase(dockerInterface DockerInterface, seedArgs ...string) error {
	output, err := runSeedData(dockerInterface, seedArgs...)
	fmt.Print(output)
	if err == nil {
		return nil
	}

	if len(seedArgs) > 0 && seedArgsUnsupported(output, seedArgs...) {
		fmt.Println("[!] This Ghostwriter build does not support the requested seed option(s); retrying without them...")
		output, retryErr := runSeedData(dockerInterface)
		fmt.Print(output)
		if retryErr == nil {
			return nil
		}
		return fmt.Errorf("optional seed arguments are unsupported, and retrying without them failed: %w", retryErr)
	}

	return err
}

// runSeedData builds and runs the docker compose command for /seed_data.
func runSeedData(dockerInterface DockerInterface, seedArgs ...string) (string, error) {
	seedCmd := append([]string{"run", "--rm", "django", "/seed_data"}, seedArgs...)
	return dockerInterface.RunComposeCmdWithCombinedOutput(seedCmd...)
}

// seedArgsUnsupported reports whether Django's loaddata command rejected one
// of the optional seed arguments. This matches the observed invalid-argument
// output from Ghostwriter's local loaddata command:
//
//	manage.py loaddata: error: unrecognized arguments: --unknown-option
func seedArgsUnsupported(output string, seedArgs ...string) bool {
	lowerOutput := strings.ToLower(output)
	loaddataInvalidArgs := "manage.py loaddata: error: unrecognized arguments:"
	if !strings.Contains(lowerOutput, loaddataInvalidArgs) {
		return false
	}

	for _, seedArg := range seedArgs {
		if strings.Contains(lowerOutput, strings.ToLower(seedArg)) {
			return true
		}
	}

	return false
}
