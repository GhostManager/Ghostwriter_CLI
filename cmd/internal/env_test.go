package internal

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"strings"
	"testing"
)

func TestGhostwriterEnvironmentVariables(t *testing.T) {
	defer quietTests()()

	// Test parsing values and writing to the .env file
	envFile := filepath.Join(GetCwdFromExe(), ".env")
	ParseGhostwriterEnvironmentVariables()
	assert.True(t, FileExists(envFile), "Expected .env file to exist")

	// Test a default value
	assert.Equal(t, ghostEnv.Get("django_date_format"), "d M Y", "Value of `django_date_format` should be `d M Y`")

	// Test modifying the .env file for production mode
	SetProductionMode()
	assert.Equal(t, ghostEnv.GetBool("hasura_graphql_dev_mode"), false, "Production value of `hasura_graphql_dev_mode` should be false")
	assert.Equal(t, ghostEnv.GetBool("django_secure_ssl_redirect"), true, "Production value of `django_secure_ssl_redirect` should be true")
	assert.Equal(t, ghostEnv.GetString("django_settings_module"), "config.settings.production", "Production value of `django_settings_module` should be `config.settings.production`")

	// Test modifying the .env file for dev mode
	SetDevMode()
	assert.Equal(t, ghostEnv.GetBool("hasura_graphql_dev_mode"), true, "Development value of `hasura_graphql_dev_mode` should be true")
	assert.Equal(t, ghostEnv.GetBool("django_secure_ssl_redirect"), false, "Development value of `django_secure_ssl_redirect` should be false")
	assert.Equal(t, ghostEnv.GetString("django_settings_module"), "config.settings.local", "Development value of `django_settings_module` should be `config.settings.local`")

	// Test ``GetConfig()``
	format := GetConfig([]string{"django_compress_enabled", "django_date_format"})
	assert.Equal(
		t,
		format,
		Configurations{Configuration{Key: "django_compress_enabled", Val: "true"}, Configuration{Key: "django_date_format", Val: "d M Y"}},
		"`GetConfig()` should return a Configurations object",
	)
	assert.Equal(t, len(format), 2, "`GetConfig()` with two valid variables should return a two values")

	// Test ``GetConfigAll()``
	config := GetConfigAll()
	assert.Equal(t, len(config), 62, "`GetConfigAll()` should return all values")

	// Test ``SetConfig()``
	SetConfig("django_date_format", "Y M d")
	assert.Equal(t, ghostEnv.GetString("django_date_format"), "Y M d", "Default value of `django_date_format` should be `Y M d`")

	// Test ``AllowHost()``
	AllowHost("test.local")
	assert.True(t, strings.Contains(ghostEnv.GetString("django_allowed_hosts"), "test.local"), "Value of `django_allowed_hosts` should include `test.local`")
	AllowHost("test.local")
	assert.Equal(t, strings.Count(ghostEnv.GetString("django_allowed_hosts"), "test.local"), 1, "Value of `django_allowed_hosts` should include only one entry for `test.local`")
	AllowHost("ghostwriter.local")
	assert.True(t, strings.Contains(ghostEnv.GetString("django_allowed_hosts"), "ghostwriter.local"), "Value of `django_allowed_hosts` should include `ghostwriter.local`")

	// Test ``DisallowHosts()``
	DisallowHost("test.local")
	assert.False(t, strings.Contains(ghostEnv.GetString("django_allowed_hosts"), "test.local"), "Value of `django_allowed_hosts` should no longer include `test.local`")
	assert.True(t, strings.Contains(ghostEnv.GetString("django_allowed_hosts"), "ghostwriter.local"), "Value of `django_allowed_hosts` should still include `ghostwriter.local`")
	DisallowHost("ghostwriter.local")
	assert.False(t, strings.Contains(ghostEnv.GetString("django_allowed_hosts"), "ghostwriter.local"), "Value of `django_allowed_hosts` should no longer include `ghostwriter.local`")

	// Test ``TrustOrigin()``
	TrustOrigin("test.local")
	assert.True(t, strings.Contains(ghostEnv.GetString("django_csrf_trusted_origins"), "test.local"), "Value of `django_csrf_trusted_origins` should include `test.local`")
	TrustOrigin("test.local")
	assert.Equal(t, strings.Count(ghostEnv.GetString("django_csrf_trusted_origins"), "test.local"), 1, "Value of `django_csrf_trusted_origins` should include only one entry for `test.local`")
	TrustOrigin("ghostwriter.local")
	assert.True(t, strings.Contains(ghostEnv.GetString("django_csrf_trusted_origins"), "ghostwriter.local"), "Value of `django_csrf_trusted_origins` should include `ghostwriter.local`")

	// Test ``DistrustOrigin()``
	DistrustOrigin("test.local")
	assert.False(t, strings.Contains(ghostEnv.GetString("django_csrf_trusted_origins"), "test.local"), "Value of `django_csrf_trusted_origins` should include `test.local`")
	assert.True(t, strings.Contains(ghostEnv.GetString("django_csrf_trusted_origins"), "ghostwriter.local"), "Value of `django_csrf_trusted_origins` should still include `ghostwriter.local`")
	DistrustOrigin("ghostwriter.local")
	assert.False(t, strings.Contains(ghostEnv.GetString("django_csrf_trusted_origins"), "ghostwriter.local"), "Value of `django_csrf_trusted_origins` should include `ghostwriter.local`")
}
