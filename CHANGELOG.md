# Changelog

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
