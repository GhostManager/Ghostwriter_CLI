# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0-rc1] - 2025-02-13

### Added

* Added the spaCy model selection to the configuration options for Ghostwriter v6.3.0
* Added a _settings/_ directory in the Ghostwriter data directory for custom Django configuration files
  * This directory allows users running published container images (prod mode) to customize Django settings without needing a local codebase
  * The directory includes a comprehensive README.md with examples for common customizations like SSO and email configuration
  * Settings files are loaded in alphabetical order, allowing number prefixes for controlled loading sequence (e.g., `1-sso-config.py`)
  * Only applies to the default production mode; `local-dev` and `local-prod` modes should continue using the codebase's _config/settings/production.d_ and _local.d_ directories
* Added `migrate` command to help users transition from legacy Ghostwriter installations to the new XDG data directory structure
  * Migrates SSL certificates from _ssl/_ to the data directory's _ssl/_ subdirectory
  * Migrates _.env_ configuration file to the data directory
  * Migrates custom Django settings from _config/settings/production.d/_ to the data directory's _settings/_ subdirectory
  * Automatically detects and migrates Docker volumes from old installations
    * Migrates PostgreSQL database (_production\_postgres\_data_)
    * Migrates media files and user uploads (_production\_data_)
    * Migrates backup archives (_production\_postgres\_data\_backups_)
    * Static files are regenerated automatically (no migration needed)
  * Creates safety backup before volume migration to prevent data loss
  * Verifies file counts after volume migration to ensure successful copy
  * Offers to clean up old volumes after successful migration to free disk space
  * Prompts for confirmation before overwriting existing files
  * Source files and volumes remain in place (copy, not move) for safe verification
  * Run from your old Ghostwriter directory with: `ghostwriter-cli migrate`

### Changed

* Changed production installations to pull published container images
  * Images are now pre-built to remove the potential for builds to fail due to issues with system resources, package managers, or DNS
    * New images are published whenever we tag a new release of Ghostwriter
  * It is no longer necessary to clone or maintain a local copy of the Ghostwriter code repository
  * New installs now take ~5 minutes, depending on available resources and network speed, a decrease of at least 4x
  * The `.env` config file and Docker YAML files now live in the system's XDG data file directory to allow Ghostwriter CLI to work from any location
    * Linux: _~/.local/share/ghostwriter/_
    * macOS: _~/Library/Application Support/ghostwriter/_
    * Windows: _%LOCALAPPDATA%/ghostwriter/_
* As part of the above change, Ghostwriter CLI now downloads the latest Docker YAML file from Ghostwriter's releases
  * This ensures you always have the file that matches your release in case we make changes for a release
* Some commands now offer `--version` flag that you can to specify the version to which you wish to update
* The `update` command now pulls the latest published container images
* The `version` command now lists the following version information:
  * The local version of Ghostwriter CLI
  * The latest release of Ghostwriter CLI available from GitHub releases
  * The version of the local Ghostwriter container images
  * The version of the latest published Ghostwriter container images
* Replaced the `--dev` flag with the `--mode` option to specify `prod` (default) or `local-prod` or `local-dev`
  * The `local-dev` and `local-prod` modes work like the old modes by using a local copy of the Ghostwriter code repository

## [0.3.0] - 2025-11-14

### Changed

* Updated the Docker client to the latest to ensure continued compatibility with Docker v29 and later (Closes #23; PR #24)
  * Docker v29 deprecated support for the Docker API v1.41 and earlier
  * This does not impact most commands for containers (e.g., `up`, `build`), but a few utility commands (e.g., `healthcheck`) used older API calls for container information

### Removed

* Removed support for the deprecated `docker-compose` v1 script (Closes #16)
  * Docker has deprecated this version and marked it as end of life as of July 2022
  * Some current (e.g., `pg-upgrade`) and future features of Ghostwriter CLI cannot support v1, so it is time to remove it to avoid confusion
  * All users should be updated to at least v2.x
  * If someone needs v1 and absolutely cannot upgrade, support for v1 remains in older Ghostwriter CLI binaries available from past releases

## [0.2.30] - 2025-11-10

### Changed

* The `backup` command now also backs up all media files in the `ghostwriter_*_data` media volumes
  * Media backups are stored alongside the database backups in the `ghostwriter_*_postgres_data_backups` volumes
* The `restore` command now includes an optional `--media` flag to restore a media backup
  * Restoring media will wipe the media volume and then replace the wiped contents with the backup
  * Both backup files will share the same timestamp in the filename for easy identification

## [0.2.29] - 2025-10-23

### Added

* Added a `migrate_totp` command to migrate TOTP secrets following upgrades in Ghostwriter v6.1
  * This is necessary to support existing users' 2FA TOTP tokens after upgrading Django's `allauth` module
* Added support for Podman as an alternative to Docker
  * Podman must be configured to use [Docker compatibility mode](https://podman-desktop.io/docs/migrating-from-docker/managing-docker-compatibility)

## [0.2.28] - 2025-07-30

### Added

* Added the `frontend` and `collab` containers to the `logs` and `running` commands for Ghostwriter v6.0.0+

## [0.2.27] - 2025-07-25

### Added

* Added `POSTGRES_CONN_MAX_AGE` to the environment variables for Ghostwriter v6.0.1+
  * This variable sets the maximum age of a PostgreSQL connection age in seconds (default: 0 seconds / disabled for ASGI)

### Changed

* Updated date formatting to make build dates from the GitHub API better match those baked into the binaries

## [0.2.26] - 2025-07-22

### Changed

* The version command now pulls the latest stable release version information for comparison and provides a download link

## [0.2.25] - 2025-07-21

### Added

* Added a `tagcleanup` command for Ghostwriter v6
  * This command runs the Django management commands to remove orphaned tags and perform tag deduplication
* Added alias commands for `containers up` and `containers down` to make it easier to run `up` and `down`

### Changed

* The `version` command now pulls the latest stable release version information for comparison and provides a download link 

## [0.2.24] - 2025-06-25

### Changed

* Updated the `pg-upgrade` command to pull the Docker network name for cases where users have customized it

## [0.2.23] - 2025-01-31

### Added

* Added a `--volumes` flag to the `containers down` command that deletes the data volumes when the containers come down
* Added an `uninstall` command that removes the target environment by deleting containers, images, and volume data

## [0.2.22] - 2025-01-08

### Fixed

* Fixed an issue with some Docker Compose commands causing a "no TTY" error during builds

## [0.2.21] - 2025-01-03

### Added

* Added `HASURA_GRAPHQL_SERVER_HOSTNAME` to the env vars for Ghostwriter v4.3.10+
* Added a `pg_upgrade` command for upgrading the PostgreSQL database (Closes #11)

## [0.2.20] - 2024-08-15

### Added

* Added configuration options for the new SSO options coming to Ghostwriter
  * Options are `django_social_account_domain_allowlist` and `django_social_account_login_on_get`

## [0.2.19] - 2024-03-29

### Changed

* Increased the time the `install` command waits for Django to start to 120 seconds

## [0.2.18] - 2024-01-12

### Added

* Added a `gencert` command to generate new SSL/TLS certificate files

### Changed

* Added a common name to the SSL/TLS certificate template

## [0.2.17] - 2023-12-14

### Changed

* Added logic to determine if the Django container failed to start when running the `install` command

## [0.2.16] - 2023-09-21

### Changed

* The `install` command now sets the initial admin user's role to `admin` for Ghostwriter v4.0.0

## [0.2.15] - 2023-08-18

### Added

* Added a configuration option for `2FA_ALWAYS_REVEAL_BACKUP_TOKENS` to control if 2FA backup codes are revealed automatically when viewing the page

## [0.2.14] - 2023-08-18

### Changed

* Changed the defaults of the session management settings to improve session security
  * Previous defaults, Django's default values, were very generous with session expiration
  * `DJANGO_SESSION_COOKIE_AGE` : Now 32,400 seconds (9 hours) down from 1,209,600 seconds (2 weeks)
  * `DJANGO_SESSION_EXPIRE_AT_BROWSER_CLOSE` : Now `true` to end sessions when the browser is closed
  * `DJANGO_SESSION_SAVE_EVERY_REQUEST` : Now `true` to automatically renew the session expiration while Ghostwriter is in-use

## [0.2.14] - 2023-08-17

### Added

* Added configuration options for managing Ghostwriter's browser sessions
  * `DJANGO_SESSION_COOKIE_AGE` sets the number of seconds a session cookie will last before expiring (default: 1209600)
  * `DJANGO_SESSION_EXPIRE_AT_BROWSER_CLOSE` sets whether the session cookie will expire when the browser is closed (default: false)
  * `DJANGO_SESSION_SAVE_EVERY_REQUEST` sets whether the session cookie will be saved on every request (default: false)

### Changed

* Commands that use Docker will now check to ensure the Docker Engine is running before proceeding

## [0.2.13] - 2023-05-23

### Changed

* Improved the help information for the `restore` command to make usage clear

## [0.2.12] - 2023-05-23

### Added

* Added `backup` and `restore` commands to make it easier to run these PostgreSQL commands to create, list, and restore database backups 

## [0.2.11] - 2023-03-27

### Added

* Added `DJANGO_SOCIAL_ACCOUNT_ALLOW_REGISTRATION` to the configuration to support changes in Ghostwriter v3.2.3

## [0.2.10] - 2023-02-21

### Security

* Upgraded the `golang.org/x/net` package to v0.7.0 to address potential security issues in <0.7.0 (CVE-2022-41723, CVE-2022-27664)

## [0.2.9] - 2023-02-10

### Changed

* Packed macOS binaries do not work with macOS v13 at this time, so macOS binaries are no longer packed with `upx --brute`
  * Only affect is macOS binaries will be slightly larger
  * Once UPX works with macOS v13, the project will return to packing the binaries to reduce their size

## [0.2.8] - 2023-02-10

### Added

* Added a `healthcheck` command that validates all Ghostwriter services are running and passing their respective health checks

### Changed

* Changed the `running` command to filter out exited containers
* Container commands will now default to using the `docker compose` command used by new installations of Compose
  * Commands will try to fall back to the deprecated `docker-compose` script when the `docker compose` plugin is not in the PATH

## [0.2.7] - 2022-10-31

### Added

* Added `--skip-seed` flag to the `containers build` subcommand to allow the database seeding step to be skipped (see below)

### Changed

* Changed the `containers build` command to (re-)seed the database with the default data (_initial.json_ files) in case values were added or adjusted in a Ghostwriter release

## [0.2.6] - 2022-10-14

### Fixed

* Changed config variable name, `HEALTHCHECK_START_PERIOD`, to the correct `HEALTHCHECK_START`

## [0.2.5] - 2022-09-30

### Added

* Added configuration options for Docker health check commands in Ghostwriter v3.0.6+

### Fixed

* Fixed Redis container missing from `running` command output for dev installations

## [0.2.4] - 2022-09-12

### Fixed

* Fixed bug where `containers restart` would restart containers using the wrong YAML file (thanks to @jenic! PR #5)

## [0.2.3] - 2022-08-05

### Fixed

* Fixed a situation that could cause Django to fail to start when switching between development and production environments

### Changed

* The Hasura Console will now remain enabled or disabled based on user preference when switching between development and production environments

## [0.2.2] - 2022-07-14

### Added

* Added two new subcommands under the `config` command: `trustorigin` and `distrustorigin`

### Changed

* DotEnv values are now wrapped in single quotes to allow for potential JSON values in the future

## [0.2.1] - 2022-06-22

### Changed

* Minor changes to help messages and information displays

## [0.2.0] - 2022-06-17

### Added

* Added `--lines` flag for the `logs` command to set the number of lines returned (default is 500)
* Added `--dev` flag to management commands to target the development environment

### Changed

* Refactored code to use Cobra CLI to improve help menus and command organization
* All commands will now default to targeting the production environment (use the `--dev` flag for development work)

## [0.1.2] - 2022-06-16

### Changed

* Removed `openssl` requirement for generating TLS/SSL certificates and Diffie-Hellman parameters
* Added `prod` as an alias for `production` so both are recognized as the same environment for commands
* Added `all` option to the `logs` command to fetch logs from all Ghostwriter containers with one command for troubleshooting

## [0.1.1] - 2022-06-07

### Changed

* Restricted default password generation to alphanumeric characters to make them safe for PostgreSQL connection strings and other services

## [0.1.0] - 2022-06-07

### Added

* First official multi-platform release for Ghostwriter v3.0.0

## [0.0.3] - 2022-06-06

### Added

* Added `allowhost` and `disallowhost` subcommands to `config` to manage the server's allowed hosts
* Added aliases for some common configurations to make it easier to retrieve the values

### Changed

* Install command will now display superuser information at the end of the install process

## [0.0.2] - 2022-06-02

### Changed

* Failure to create a Django superuser will no longer prevent the `install` command from continuing

## [0.0.1] - 2022-05-09

### Added

* Initial commit & release

### Changed

* N/A

### Deprecated

* N/A

### Removed

* N/A

### Fixed

* N/A

### Security

* N/A
