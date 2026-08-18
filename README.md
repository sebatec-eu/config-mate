# config-mate

config-mate lets you write Go apps that run unchanged on Hostsharing,
your own server, or locally. It handles the differences between FastCGI
(Fast Common Gateway Interface) and HTTP, config file locations, and logging—
so you don’t have to.

Use config-mate if you deploy Go apps to multiple environments and want one
codebase. We assume you have at least one production environment (Hostsharing
and/or VM/Root) and one development environment.

## How-to guides

### Deploy to Hostsharing

Hostsharing runs apps via Apache and FastCGI.

**Steps**

1. Upload your binary to `~/doms/<domain>/fastcgi-ssl/`
2. Ensure it has execute permissions

config-mate detects FastCGI, loads config from your domain directory, and
writes logs to your domain log directory.

### Deploy to a Root Server

Run your app on a VM or bare metal server with HTTP.

**Steps**

1. Set `ADDR` or `PORT` environment variable
2. Run your binary

config-mate runs in HTTP mode, loads config from XDG directories
(standard Linux config locations), and logs to stdout.

**Environment variables**

- `ADDR`: Listen address (e.g., `:8080`)
- `PORT`: Listen port (e.g., `8080`)

### Support Both Hostsharing and Root Server

Maintain one codebase for both platforms.

**Steps**

1. Use the same binary for both platforms

config-mate auto-detects the environment at startup and adapts its behavior.

### Develop with Vite Proxy

Run a Go backend with a Vite frontend in development.

**Steps**

1. Set `PORT` to match your Vite proxy target

Works in HTTP mode with local config.

### Use with Caddy or Apache

Run behind a reverse proxy.

**Steps**

1. For FastCGI: Set `FCGI_LISTEN`
2. For HTTP: Set `ADDR` or `PORT`

Adapts to your proxy mode.

## Testing

Run all tests:

```sh
make test
```

## See also

- [Hostsharing documentation](https://www.hostsharing.net/doc/managed-operations-platform/)
  for directory structure details
- [pkg.go.dev/github.com/sebatec-eu/config-mate](https://pkg.go.dev/github.com/sebatec-eu/config-mate)
  for API reference
- `go doc` for package documentation
