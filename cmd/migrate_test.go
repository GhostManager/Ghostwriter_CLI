package cmd

import "testing"

func TestFindLegacyVolumeByKey_MatchesCanonicalSuffix(t *testing.T) {
	volumes := []string{
		"ghostwriter_production_postgres_data",
	}

	got := findLegacyVolumeByKey(volumes, "production_postgres_data")
	if got != "ghostwriter_production_postgres_data" {
		t.Fatalf("expected canonical suffix match, got %q", got)
	}
}

func TestFindLegacyVolumeByKey_DoesNotFallbackToContains(t *testing.T) {
	volumes := []string{
		"legacy-production_postgres_data-archive",
	}

	got := findLegacyVolumeByKey(volumes, "production_postgres_data")
	if got != "" {
		t.Fatalf("expected no match for non-canonical name, got %q", got)
	}
}

func TestFindLegacyVolumeByKey_NoMatch(t *testing.T) {
	volumes := []string{"unrelated_volume"}

	got := findLegacyVolumeByKey(volumes, "production_data")
	if got != "" {
		t.Fatalf("expected no match, got %q", got)
	}
}
