package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGhostwriterEnvironmentVariables(t *testing.T) {
	defer quietTests()()

	tempDir, err := os.MkdirTemp("", "gwtest")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test parsing values and writing to the .env file
	envFile := filepath.Join(tempDir, ".env")
	env, err := ReadEnv(tempDir)
	assert.NoError(t, err)

	assert.True(t, FileExists(envFile), "Expected .env file to exist")

	// Test a default value
	assert.Equal(t, env.Get("django_date_format"), "d M Y", "Value of `django_date_format` should be `d M Y`")

	// Test modifying the .env file for production mode
	env.SetProd()
	env.Save()
	env, err = ReadEnv(tempDir)
	assert.NoError(t, err)
	assert.Equal(t, env.Get("hasura_graphql_dev_mode"), "false", "Production value of `hasura_graphql_dev_mode` should be false")
	assert.Equal(t, env.Get("django_secure_ssl_redirect"), "true", "Production value of `django_secure_ssl_redirect` should be true")
	assert.Equal(t, env.Get("django_settings_module"), "config.settings.production", "Production value of `django_settings_module` should be `config.settings.production`")

	// Test modifying the .env file for dev mode
	env.SetDev()
	env.Save()
	env, err = ReadEnv(tempDir)
	assert.NoError(t, err)
	assert.Equal(t, env.Get("hasura_graphql_dev_mode"), "true", "Development value of `hasura_graphql_dev_mode` should be true")
	assert.Equal(t, env.Get("django_secure_ssl_redirect"), "false", "Development value of `django_secure_ssl_redirect` should be false")
	assert.Equal(t, env.Get("django_settings_module"), "config.settings.local", "Development value of `django_settings_module` should be `config.settings.local`")

	// Test ``GetAll()``
	config := env.GetAll()
	assert.Equal(t, len(config), 67, "`GetConfigAll()` should return all values")

	// Test ``Set()``
	env.Set("django_date_format", "Y M d")
	assert.Equal(t, env.Get("django_date_format"), "Y M d", "Default value of `django_date_format` should be `Y M d`")

	// Test ``AppendHost()``
	env.AppendHost("django_allowed_hosts", "test.local")
	assert.True(t, strings.Contains(env.Get("django_allowed_hosts"), "test.local"), "Value of `django_allowed_hosts` should include `test.local`")
	env.AppendHost("django_allowed_hosts", "test.local")
	assert.Equal(t, strings.Count(env.Get("django_allowed_hosts"), "test.local"), 1, "Value of `django_allowed_hosts` should include only one entry for `test.local`")
	env.AppendHost("django_allowed_hosts", "ghostwriter.local")
	assert.True(t, strings.Contains(env.Get("django_allowed_hosts"), "ghostwriter.local"), "Value of `django_allowed_hosts` should include `ghostwriter.local`")

	// Test ``RemoveHost()``
	env.RemoveHost("django_allowed_hosts", "test.local")
	assert.False(t, strings.Contains(env.Get("django_allowed_hosts"), "test.local"), "Value of `django_allowed_hosts` should no longer include `test.local`")
	assert.True(t, strings.Contains(env.Get("django_allowed_hosts"), "ghostwriter.local"), "Value of `django_allowed_hosts` should still include `ghostwriter.local`")
	env.RemoveHost("django_allowed_hosts", "ghostwriter.local")
	assert.False(t, strings.Contains(env.Get("django_allowed_hosts"), "ghostwriter.local"), "Value of `django_allowed_hosts` should no longer include `ghostwriter.local`")
}
