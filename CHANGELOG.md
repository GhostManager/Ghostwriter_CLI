# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

* Added `allowhost` and `disallowhost` subdommands to `config` to manage the server's allowed hosts
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
