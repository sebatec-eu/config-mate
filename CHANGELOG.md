# Changelog

## v1.9.0 - 2026-06-29

- **Added**: `PAC()` method to `user` struct to return the Web-Paket prefix.
- **Added**: `Domain()` and `DomsDir()` methods to `domain` struct for domain hostname and directory path access.
- **Added**: `DomainByExecutable()` to parse domain from FastCGI binary path (like `DomainByWorkingDir`). Honors `CONFIG_BASE_PATH` first, allowing local dev to simulate Hostsharing layouts without `/home/pacs`. Accepts file paths (`…/api.fcgi`) and directories (`…/doms/example.com`).
- **Added**: `CONFIG_BASE_PATH` env var for local development. Set to an absolute path encoding PAC/user/dom layout to make `ConfigDir()`, `LogDir()`, and `DataDir()` resolve predictably.
- **Updated**: `hostsharing.ReadInConfig` and `database.getDataDirResolver` now use `DomainByExecutable()` instead of `DomainByWorkingDir()`, making startup robust against later `chdir`.
- **Fixed**: `hostsharing.fcgiLogFile` falls back to `os.Stdout` instead of panicking when the executable path is too shallow to extract a domain. Protects local development and `IsFCGI()` false positives.
- **Deprecated**: `DomainByWorkingDir()`. Use `DomainByExecutable()` instead. Will be removed in v2.
- **Updated**: `gorm.io/gorm` dependency (v1.31.1 → v1.31.2).
- **Updated**: `.devcontainer.json` configuration cleanup.

## v1.8.0 - 2026-06-04

- **Added**: `.agents/skills/changelog-writer` skill for structured changelog management.
- **Updated**: `chi/v5` (v5.2.3 → v5.3.0) and `httplog/v3` (v3.3.0 → v3.4.0).
- **Updated**: GitHub Actions pinned versions for `checkout` and `setup-go`.

## v1.7.0 - 2026-01-11

- **Added**: Comprehensive API documentation comments for `database` and `hostsharing` packages.
- **Added**: Test coverage for `DomainByWorkingDir()` function with 11 test cases.
- **Improved**: Configuration file loading now auto-detects app name from `ServiceName()` when not provided.
- **Updated**: Dependencies including `gorm` (v1.30.5 → v1.31.1), `httplog/v3` (v3.2.2 → v3.3.0), and SQLite driver.
- **Updated**: GitHub Actions to use latest versions (checkout v6, setup-go v6).

## v1.6.0 - 2025-09-12

- **Added**: Generic database support for **SQLite** and **MySQL** (via `gorm`).
- **Added**: Automatic **service name detection** via the `SERVICE_NAME` environment variable or executable name.
- **Added**: Flexible server startup with **FastCGI** (via `FCGI_LISTEN` environment variable) or **HTTP** (default port `9000`).
- **Improved**: Default SQLite database name (`./data.db` or dynamically in the `DataDir` of the Hostsharing domain).

## v1.5.0 - 2025-09-11

- **Added**: Generic database package with support for **SQLite3** and **MySQL/MariaDB**.

## v1.4.0 - 2025-09-09

- **Added**: Extended logging functionality.
- **Added**: CSRF middleware for HTTP requests.
- **Updated**: Migration to `httplog/v3`.

## v1.3.0 - 2025-06-25

- **Added**: Support for **SVG files** in the static file handler.
- **Added**: Caching for static files.
- **Added**: JSON support in the `combined_handler`.

## v1.2.0 - 2025-06-23

- **Added**: Compression middleware with customizable defaults.

## v1.1.0 - 2025-06-20

- **Added**: More flexible app name detection.
- **Added**: FCGI request logger with auto-detection.
- **Fixed**: Configuration reader corrections.

## v1.0.0 - 2025-06-19

- **Initial Release**: Basic configuration management and UI handler.
