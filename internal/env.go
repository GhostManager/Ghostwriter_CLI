package internal

// Functions for managing the environment variables that control the
// configuration of the Ghostwriter containers.

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var ghostEnv = viper.New()

// Set sane defaults for a basic Ghostwriter deployment.
// Defaults are geared towards a development environment.
func setGhostwriterConfigDefaultValues() {
	// Project configuration
	ghostEnv.SetDefault("use_docker", "yes")
	ghostEnv.SetDefault("ipythondir", "/app/.ipython")

	// Django configuration
	ghostEnv.SetDefault("django_account_allow_registration", false)
	ghostEnv.SetDefault("django_account_email_verification", "none")
	ghostEnv.SetDefault("django_admin_url", "admin/")
	ghostEnv.SetDefault("django_allowed_hosts", "localhost 127.0.0.1 172.20.0.5 django host.docker.internal ghostwriter.local")
	ghostEnv.SetDefault("django_compress_enabled", true)
	ghostEnv.SetDefault("django_date_format", "d M Y")
	ghostEnv.SetDefault("django_host", "django")
	ghostEnv.SetDefault("django_jwt_secret_key", GenerateRandomPassword(32, false))
	ghostEnv.SetDefault("django_mailgun_api_key", "")
	ghostEnv.SetDefault("django_mailgun_domain", "")
	ghostEnv.SetDefault("django_port", "8000")
	ghostEnv.SetDefault("django_qcluster_name", "soar")
	ghostEnv.SetDefault("django_secret_key", GenerateRandomPassword(32, false))
	ghostEnv.SetDefault("django_secure_ssl_redirect", false)
	ghostEnv.SetDefault("django_settings_module", "config.settings.local")
	ghostEnv.SetDefault("django_superuser_email", "admin@ghostwriter.local")
	ghostEnv.SetDefault("django_superuser_password", GenerateRandomPassword(32, true))
	ghostEnv.SetDefault("django_superuser_username", "admin")
	ghostEnv.SetDefault("django_web_concurrency", 4)

	// PostgreSQL configuration
	ghostEnv.SetDefault("postgres_host", "postgres")
	ghostEnv.SetDefault("postgres_port", 5432)
	ghostEnv.SetDefault("postgres_db", "ghostwriter")
	ghostEnv.SetDefault("postgres_user", "postgres")
	ghostEnv.SetDefault("postgres_password", GenerateRandomPassword(32, true))

	// Redis configuration
	ghostEnv.SetDefault("redis_host", "redis")
	ghostEnv.SetDefault("redis_port", 6379)

	// Nginx configuration
	ghostEnv.SetDefault("nginx_host", "nginx")
	ghostEnv.SetDefault("nginx_port", 443)

	// Hasura configuration
	ghostEnv.SetDefault("hasura_graphql_action_secret", GenerateRandomPassword(32, true))
	ghostEnv.SetDefault("hasura_graphql_admin_secret", GenerateRandomPassword(32, true))
	ghostEnv.SetDefault("hasura_graphql_dev_mode", true)
	ghostEnv.SetDefault("hasura_graphql_enable_console", true)
	ghostEnv.SetDefault("hasura_graphql_enabled_log_types", "startup, http-log, webhook-log, websocket-log, query-log")
	ghostEnv.SetDefault("hasura_graphql_enable_telemetry", false)
	ghostEnv.SetDefault("hasura_graphql_server_host", "graphql_engine")
	ghostEnv.SetDefault("hasura_graphql_insecure_skip_tls_verify", true)
	ghostEnv.SetDefault("hasura_graphql_log_level", "warn")
	ghostEnv.SetDefault("hasura_graphql_metadata_dir", "/metadata")
	ghostEnv.SetDefault("hasura_graphql_migrations_dir", "/migrations")
	ghostEnv.SetDefault("hasura_graphql_server_port", 8080)

	// Set some elpful aliases for common settings
	ghostEnv.RegisterAlias("date_format", "django_date_format")
	ghostEnv.RegisterAlias("admin_password", "django_superuser_password")
	ghostEnv.RegisterAlias("hasura_password", "hasura_graphql_admin_secret")
}

// Write the environment variables to the ``.env`` file.
func WriteGhostwriterEnvironmentVariables() {
	c := ghostEnv.AllSettings()
	// To make it easier to read and look at, get all the keys, sort them, and display variables in order
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	f, err := os.Create(filepath.Join(GetCwdFromExe(), ".env"))
	if err != nil {
		log.Fatalf("Error writing out environment!\n%v", err)
	}
	defer f.Close()
	for _, key := range keys {
		if len(ghostEnv.GetString(key)) == 0 {
			_, err = f.WriteString(fmt.Sprintf("%s=\n", strings.ToUpper(key)))
		} else {
			_, err = f.WriteString(fmt.Sprintf("%s=\"%s\"\n", strings.ToUpper(key), ghostEnv.GetString(key)))
		}

		if err != nil {
			log.Fatalf("Failed to write out environment!\n%v", err)
		}
	}
}

// Attempt to find and open an existing .env file or create a new one.
// If an .env file is found, load it into the Viper configuration.
// If an .env file is not found, create a new one with default values.
// Then write the final file with ``WriteGhostwriterEnvironmentVariables()``.
func ParseGhostwriterEnvironmentVariables() {
	setGhostwriterConfigDefaultValues()
	ghostEnv.SetConfigName(".env")
	ghostEnv.SetConfigType("env")
	ghostEnv.AddConfigPath(".")
	ghostEnv.AutomaticEnv()
	// Check if expected env file exists
	if !FileExists(filepath.Join(GetCwdFromExe(), ".env")) {
		_, err := os.Create(filepath.Join(GetCwdFromExe(), ".env"))
		if err != nil {
			log.Fatalf("The .env doesn't exist and couldn't be created")
		}
	}
	// Try reading the env file
	if err := ghostEnv.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("Error while reading in .env file: %s", err)
		} else {
			log.Fatalf("Error while parsing .env file: %s", err)
		}
	}
	WriteGhostwriterEnvironmentVariables()
}

// Update the environment variables to switch to production mode.
func SetProductionMode() {
	ghostEnv.Set("hasura_graphql_dev_mode", false)
	ghostEnv.Set("django_secure_ssl_redirect", true)
	ghostEnv.Set("hasura_graphql_insecure_skip_tls_verify", false)
	ghostEnv.Set("django_settings_module", "config.settings.production")
	WriteGhostwriterEnvironmentVariables()
}

// Update the environment variables to switch to development mode.
func SetDevMode() {
	ghostEnv.Set("hasura_graphql_dev_mode", true)
	ghostEnv.Set("django_secure_ssl_redirect", false)
	ghostEnv.Set("hasura_graphql_insecure_skip_tls_verify", true)
	ghostEnv.Set("django_settings_module", "config.settings.local")
	WriteGhostwriterEnvironmentVariables()
}

// Update the environment variables to allow a new hostname.
func AppendAllowedHost(host string) {
	current := ghostEnv.GetString("django_allowed_hosts")
	updated_string := fmt.Sprintf("%s %s", current, host)
	ghostEnv.Set("django_allowed_hosts", updated_string)
}

// Update the environment variables to disallow a hostname.
func RemoveAllowedHost(host string) {
	current := ghostEnv.GetString("django_allowed_hosts")
	current = strings.Replace(current, host, "", 1)
	current = strings.Replace(current, "  ", " ", 1)
	ghostEnv.Set("django_allowed_hosts", strings.TrimSpace(current))
}

// Review, get, or set environment variables.
// Prints contents of the .env if no arguments are provided.
func Env(args []string) {
	if len(args) == 0 {
		fmt.Println("[+] Current configuration and available variables:")
		c := ghostEnv.AllSettings()
		keys := make([]string, 0, len(c))
		for k := range c {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Println(strings.ToUpper(key), "=", ghostEnv.Get(key))
		}
		return
	}

	switch args[0] {
	case "get":
		if len(args) == 1 {
			log.Fatal("Must specify name of variable to get")
		}
		for i := 1; i < len(args[1:])+1; i++ {
			setting := strings.ToLower(args[i])
			val := ghostEnv.Get(setting)
			if val == nil {
				log.Fatalf("Config variable `%s` not found", setting)
			} else {
				fmt.Printf("\n%s: %s\n", strings.ToUpper(setting), val)
			}
		}
	case "set":
		if len(args) != 3 {
			log.Fatalf("Must supply config name and config value")
		}
		if strings.ToLower(args[2]) == "true" {
			ghostEnv.Set(args[1], true)
		} else if strings.ToLower(args[2]) == "false" {
			ghostEnv.Set(args[1], false)
		} else {
			ghostEnv.Set(args[1], args[2])
		}
		ghostEnv.Get(args[1])
		WriteGhostwriterEnvironmentVariables()
		fmt.Printf("[+] Successfully updated configuration in .env\n")
	case "allowhost":
		if len(args) != 2 {
			log.Fatalf("Must supply config name and config value")
		}
		new_host := strings.ToLower(args[1])
		AppendAllowedHost(new_host)
		WriteGhostwriterEnvironmentVariables()
		fmt.Printf("[+] Successfully updated configuration in .env\n")
	case "disallowhost":
		if len(args) != 2 {
			log.Fatalf("Must supply config name and config value")
		}
		host := strings.ToLower(args[1])
		RemoveAllowedHost(host)
		WriteGhostwriterEnvironmentVariables()
		fmt.Printf("[+] Successfully updated configuration in .env\n")

	default:
		fmt.Println("[-] Unknown env subcommand:", args[0])
	}
}
