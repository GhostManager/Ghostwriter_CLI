package internal

// Functions for managing the environment variables that control the
// configuration of the Ghostwriter containers.

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/viper"
)

type GWEnvironment struct {
	filepath string
	env      *viper.Viper
}

func ReadEnv(dir string) (*GWEnvironment, error) {
	filepath := filepath.Join(dir, ".env")
	// Create empty file if it doesn't exist
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}

	env := viper.New()
	env.SetConfigType("env")
	env.SetConfigFile(filepath)
	env.AutomaticEnv()
	err = env.ReadInConfig()
	if err != nil {
		return nil, err
	}

	setDefaultConfigValues(env)

	return &GWEnvironment{filepath: filepath, env: env}, nil
}

func (this *GWEnvironment) Save() {
	// Viper's write does not sort keys, so implement our own that does.
	// Use the write-and-rename pattern to atomically update.

	// Preserve existing file permissions, or use 0600 as default
	perm := os.FileMode(0600)
	if info, err := os.Stat(this.filepath); err == nil {
		perm = info.Mode().Perm()
	}

	dir := filepath.Dir(this.filepath)
	file, err := os.CreateTemp(dir, ".env")
	if err != nil {
		log.Fatalf("Could not create environmental variables file: %s\n", err)
	}

	entries := this.GetAll()
	for _, entry := range entries {
		if len(entry.Val) == 0 {
			_, err = fmt.Fprintf(file, "%s=\n", strings.ToUpper(entry.Key))
		} else {
			_, err = fmt.Fprintf(file, "%s='%s'\n", strings.ToUpper(entry.Key), entry.Val)
		}
		if err != nil {
			file.Close()
			os.Remove(file.Name())
			log.Fatalf("Could not write to environmental variables file: %s\n", err)
		}
	}

	file.Sync()
	err = file.Close()
	if err != nil {
		os.Remove(file.Name())
		log.Fatalf("Could not write to environmental variables file: %s\n", err)
	}

	err = os.Rename(file.Name(), this.filepath)
	if err != nil {
		os.Remove(file.Name())
		log.Fatalf("Could not save environmental variables: %s\n", err)
	}

	// Apply preserved permissions
	err = os.Chmod(this.filepath, perm)
	if err != nil {
		log.Fatalf("Could not set permissions on environmental variables file: %s\n", err)
	}
}

func (this *GWEnvironment) SetDev() {
	this.env.Set("hasura_graphql_dev_mode", true)
	this.env.Set("django_secure_ssl_redirect", false)
	this.env.Set("django_settings_module", "config.settings.local")
	this.env.Set("django_csrf_cookie_secure", false)
	this.env.Set("django_session_cookie_secure", false)
}

func (this *GWEnvironment) SetProd() {
	this.env.Set("hasura_graphql_dev_mode", false)
	this.env.Set("django_secure_ssl_redirect", true)
	this.env.Set("django_settings_module", "config.settings.production")
	this.env.Set("django_csrf_cookie_secure", true)
	this.env.Set("django_session_cookie_secure", true)
}

func (this *GWEnvironment) Get(key string) string {
	return this.env.GetString(key)
}

func (this *GWEnvironment) Set(key string, val string) {
	this.env.Set(key, val)
}

func (this *GWEnvironment) AppendHost(key string, host string) {
	value := this.Get(key)
	values := strings.Split(value, " ")
	if slices.Contains(values, host) {
		return
	}

	values = append(values, host)
	value = strings.Join(values, " ")
	this.Set(key, value)
}

func (this *GWEnvironment) RemoveHost(key string, host string) {
	value := this.Get(key)
	values := strings.Split(value, " ")

	newValues := []string{}
	for _, v := range values {
		if v != host {
			newValues = append(newValues, v)
		}
	}

	value = strings.Join(newValues, " ")
	this.Set(key, value)
}

func (this *GWEnvironment) GetAll() []Configuration {
	keys := this.env.AllKeys()
	slices.Sort(keys)
	out := []Configuration{}
	for _, key := range keys {
		out = append(out, Configuration{
			Key: key,
			Val: this.Get(key),
		})
	}
	return out
}

// Configuration is a custom type for storing configuration values as Key:Val pairs.
type Configuration struct {
	Key string
	Val string
}

// Set sane defaults for a basic Ghostwriter deployment.
// Defaults are geared towards a development environment.
func setDefaultConfigValues(env *viper.Viper) {
	// GW-CLI configuration
	env.SetDefault("gwcli_auto_check_updates", true)

	// Project configuration
	env.SetDefault("use_docker", "yes")
	env.SetDefault("ipythondir", "/app/.ipython")

	// Passive Voice Detection configuration
	env.SetDefault("spacy_model", "en_core_web_sm")

	// Django configuration
	env.SetDefault("django_mfa_always_reveal_backup_tokens", false)
	env.SetDefault("django_account_allow_registration", false)
	env.SetDefault("django_account_reauthentication_timeout", 32400)
	env.SetDefault("django_account_email_verification", "none")
	env.SetDefault("django_admin_url", "admin/")
	env.SetDefault("django_allowed_hosts", "localhost 127.0.0.1 django nginx host.docker.internal ghostwriter.local")
	env.SetDefault("django_compress_enabled", true)
	env.SetDefault("django_csrf_cookie_secure", false)
	env.SetDefault("django_csrf_trusted_origins", "")
	env.SetDefault("django_date_format", "d M Y")
	env.SetDefault("django_host", "django")
	env.SetDefault("django_jwt_secret_key", GenerateRandomPassword(32, false))
	env.SetDefault("django_mailgun_api_key", "")
	env.SetDefault("django_mailgun_domain", "")
	env.SetDefault("django_port", "8000")
	env.SetDefault("django_qcluster_name", "soar")
	env.SetDefault("django_secret_key", GenerateRandomPassword(32, false))
	env.SetDefault("django_secure_ssl_redirect", false)
	env.SetDefault("django_session_cookie_age", 32400)
	env.SetDefault("django_session_cookie_secure", false)
	env.SetDefault("django_session_expire_at_browser_close", false)
	env.SetDefault("django_session_save_every_request", true)
	env.SetDefault("django_settings_module", "config.settings.local")
	env.SetDefault("django_social_account_allow_registration", false)
	env.SetDefault("django_social_account_domain_allowlist", "")
	env.SetDefault("django_social_account_login_on_get", false)
	env.SetDefault("django_superuser_email", "admin@ghostwriter.local")
	env.SetDefault("django_superuser_password", GenerateRandomPassword(32, true))
	env.SetDefault("django_superuser_username", "admin")
	env.SetDefault("django_web_concurrency", 4)

	// PostgreSQL configuration
	env.SetDefault("postgres_host", "postgres")
	env.SetDefault("postgres_port", 5432)
	env.SetDefault("postgres_db", "ghostwriter")
	env.SetDefault("postgres_user", "postgres")
	env.SetDefault("postgres_password", GenerateRandomPassword(32, true))
	env.SetDefault("POSTGRES_CONN_MAX_AGE", 0)

	// Redis configuration
	env.SetDefault("redis_host", "redis")
	env.SetDefault("redis_port", 6379)

	// Nginx configuration
	env.SetDefault("nginx_host", "nginx")
	env.SetDefault("nginx_port", 443)

	// Hasura configuration
	env.SetDefault("hasura_graphql_action_secret", GenerateRandomPassword(32, true))
	env.SetDefault("hasura_graphql_admin_secret", GenerateRandomPassword(32, true))
	env.SetDefault("hasura_graphql_dev_mode", true)
	env.SetDefault("hasura_graphql_enable_console", false)
	env.SetDefault("hasura_graphql_enabled_log_types", "startup, http-log, webhook-log, websocket-log, query-log")
	env.SetDefault("hasura_graphql_enable_telemetry", false)
	env.SetDefault("hasura_graphql_server_host", "graphql_engine")
	env.SetDefault("hasura_graphql_server_hostname", "graphql_engine")
	env.SetDefault("hasura_graphql_insecure_skip_tls_verify", true)
	env.SetDefault("hasura_graphql_log_level", "warn")
	env.SetDefault("hasura_graphql_metadata_dir", "/metadata")
	env.SetDefault("hasura_graphql_migrations_dir", "/migrations")
	env.SetDefault("hasura_graphql_server_port", 8080)

	// Docker & Django health check configuration
	env.SetDefault("healthcheck_disk_usage_max", 90)
	env.SetDefault("healthcheck_interval", "300s")
	env.SetDefault("healthcheck_mem_min", 100)
	env.SetDefault("healthcheck_retries", 3)
	env.SetDefault("healthcheck_start", "60s")
	env.SetDefault("healthcheck_timeout", "30s")

	// Set some helpful aliases for common settings
	env.RegisterAlias("date_format", "django_date_format")
	env.RegisterAlias("admin_password", "django_superuser_password")
	env.RegisterAlias("hasura_password", "hasura_graphql_admin_secret")
	env.RegisterAlias("spacy", "spacy_model")
}
