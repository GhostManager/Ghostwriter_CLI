package internal

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGhostwriterEnvironmentVariables(t *testing.T) {
	defer quietTests()()

	// Test parsing values and writing to the .env file
	env_file := filepath.Join(GetCwdFromExe(), ".env")
	ParseGhostwriterEnvironmentVariables()
	if !FileExists(env_file) {
		t.Error("Expected `.env` file to be created")
	}

	// Test a default value
	if ghostEnv.Get("django_date_format") != "d M Y" {
		t.Errorf("Expected `django_date_format` to be `d M Y`, got %s", ghostEnv.Get("django_date_format"))
	}

	// Test modifying the .env file for production mode
	SetProductionMode()
	if ghostEnv.GetBool("hasura_graphql_dev_mode") {
		t.Error("Expected `hasura_graphql_dev_mode` to be false, got true")
	}
	if !ghostEnv.GetBool("django_secure_ssl_redirect") {
		t.Error("Expected `django_secure_ssl_redirect` to be true, got false")
	}
	if ghostEnv.GetBool("hasura_graphql_enable_console") {
		t.Error("Expected `hasura_graphql_enable_console` to be false, got true")
	}
	if ghostEnv.GetBool("hasura_graphql_insecure_skip_tls_verify") {
		t.Error("Expected `hasura_graphql_insecure_skip_tls_verify` to be false, got true")
	}
	if ghostEnv.Get("django_settings_module") != "config.settings.production" {
		t.Errorf("Expected `django_settings_module` to be `config.settings.production`, got %s", ghostEnv.Get("django_settings_module"))
	}

	// Test modifying the .env file for dev mode
	SetDevMode()
	if !ghostEnv.GetBool("hasura_graphql_dev_mode") {
		t.Error("Expected `hasura_graphql_dev_mode` to be true, got false")
	}
	if ghostEnv.GetBool("django_secure_ssl_redirect") {
		t.Error("Expected `django_secure_ssl_redirect` to be false, got true")
	}
	if !ghostEnv.GetBool("hasura_graphql_enable_console") {
		t.Error("Expected `hasura_graphql_enable_console` to be true, got false")
	}
	if !ghostEnv.GetBool("hasura_graphql_insecure_skip_tls_verify") {
		t.Error("Expected `hasura_graphql_insecure_skip_tls_verify` to be true, got false")
	}
	if ghostEnv.Get("django_settings_module") != "config.settings.local" {
		t.Errorf("Expected `django_settings_module` to be `config.settings.local`, got %s", ghostEnv.Get("django_settings_module"))
	}

	// Test ``Env()`` with different arguments
	Env([]string{"get", "django_date_format"})
	Env([]string{"set", "django_date_format", "Y M d"})
	if ghostEnv.Get("django_date_format") != "Y M d" {
		t.Errorf("Set `django_date_format` to `Y M d`, got %s instead", ghostEnv.Get("django_date_format"))
	}
	Env([]string{"allowhost", "test.local"})
	if !strings.Contains(ghostEnv.GetString("django_allowed_hosts"), "test.local") {
		t.Errorf("Expected `django_allowed_hosts` to contain `test.local`, got %s", ghostEnv.Get("django_allowed_hosts"))
	}
	Env([]string{"disallowhost", "test.local"})
	if strings.Contains(ghostEnv.GetString("django_allowed_hosts"), "test.local") {
		t.Errorf("Expected `django_allowed_hosts` to NOT contain `test.local`, got %s", ghostEnv.Get("django_allowed_hosts"))
	}
}
