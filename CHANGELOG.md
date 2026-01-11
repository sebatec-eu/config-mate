# Changelog

## v1.7.0

- **Added**: Comprehensive API documentation comments for `database` and `hostsharing` packages.
- **Added**: Test coverage for `DomainByWorkingDir()` function with 11 test cases.
- **Improved**: Configuration file loading now auto-detects app name from `ServiceName()` when not provided.
- **Updated**: Dependencies including `gorm` (v1.30.5 → v1.31.1), `httplog/v3` (v3.2.2 → v3.3.0), and SQLite driver.
- **Updated**: GitHub Actions to use latest versions (checkout v6, setup-go v6).

## v1.6.0

- **Added**: Generic database support for **SQLite** and **MySQL** (via `gorm`).
- **Added**: Automatic **service name detection** via the `SERVICE_NAME` environment variable or executable name.
- **Added**: Flexible server startup with **FastCGI** (via `FCGI_LISTEN` environment variable) or **HTTP** (default port `9000`).
- **Improved**: Default SQLite database name (`./data.db` or dynamically in the `DataDir` of the Hostsharing domain).

## v1.5.0

- **Added**: Generic database package with support for **SQLite3** and **PostgreSQL**.

## v1.4.0

- **Added**: Extended logging functionality.
- **Added**: CSRF middleware for HTTP requests.
- **Updated**: Migration to `httplog/v3`.

## v1.3.0

- **Added**: Support for **SVG files** in the static file handler.
- **Added**: Caching for static files.
- **Added**: JSON support in the `combined_handler`.

## v1.2.0

- **Added**: Compression middleware with customizable defaults.

## v1.1.0

- **Added**: More flexible app name detection.
- **Added**: FCGI request logger with auto-detection.
- **Fixed**: Configuration reader corrections.

## v1.0.0

- **Initial Release**: Basic configuration management and UI handler.
