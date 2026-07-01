package internal

import "testing"

func TestSeedArgsUnsupportedDetectsLoaddataInvalidArgument(t *testing.T) {
	output := `usage: manage.py loaddata [-h] [--force]
                          fixture [fixture ...]
manage.py loaddata: error: unrecognized arguments: --required-only`

	if !seedArgsUnsupported(output, "--required-only") {
		t.Fatalf("expected loaddata invalid-argument output to be detected")
	}
}

func TestSeedArgsUnsupportedIgnoresOtherParserErrors(t *testing.T) {
	output := `usage: manage.py loaddata [-h] [--force]
                          fixture [fixture ...]
manage.py loaddata: error: the following arguments are required: fixture`

	if seedArgsUnsupported(output, "--required-only") {
		t.Fatalf("expected non-invalid-argument parser error to be ignored")
	}
}

func TestSeedArgsUnsupportedIgnoresOtherSeedFailures(t *testing.T) {
	output := "django.db.utils.IntegrityError: duplicate key value violates unique constraint"

	if seedArgsUnsupported(output, "--required-only") {
		t.Fatalf("expected unrelated seed failure to be ignored")
	}
}
