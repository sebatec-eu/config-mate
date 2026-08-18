# config-mate

Scaffolding for simple Go applications that need to run in three environments
without forking the code:

- **Hostsharing** — Apache + FastCGI behind `/fastcgi-bin/` aliased to
  `~/doms/<host>/fastcgi/`.
- **VM/Root server** — systemd-style deployment behind a Caddy reverse proxy
  (HTTP, possibly FastCGI via `transport fastcgi`).
- **Local development** — `go run`, plain HTTP or Caddy + `FCGI_LISTEN`.

## Package layout

The module is split into three packages so each app type imports only what it
needs:

| Package       | Role                                                                  | Depends on            |
|---------------|-----------------------------------------------------------------------|-----------------------|
| `core/`       | Environment-agnostic helpers                                          | (nothing)             |
| `server/`     | Environment-aware server glue                                         | `core`, `hostsharing` |
| `hostsharing/`| Hostsharing-path utilities (`ParseDomain`, `DomainByExecutable`, `FcgiLogFile`) | (nothing)      |
| `database/`   | SQLite/MySQL with a `DataDirResolver` seam                            | `core`, `hostsharing` |
| `ui/`         | Static / template handler, compression                                | (nothing)             |

### Import profile by app type

| App type                                  | Imports                                                  |
|-------------------------------------------|----------------------------------------------------------|
| 1. Hostsharing only                       | `core`, `server`, `hostsharing`, `database`, `ui`        |
| 2. Hostsharing + VM/Root                  | `core`, `server`, `hostsharing`, `database`, `ui`        |
| 3. VM/Root only                           | `core`, `server`, `database`, `ui`                       |

Anything **not** in `hostsharing/` works without a Hostsharing environment.
That means a VM-only app can compile and run without ever importing
`hostsharing/`. The `database.DataDirResolverFunc` seam makes that explicit:
override it before calling `database.Open` to point SQLite at a VM-style data
directory.

## Environment-detection rules

- `core.IsFCGI()` returns true when the executable's parent directory's base
  name starts with `fastcgi`. This is the convention used by Hostsharing's
  Apache alias `/fastcgi-bin/` and by any reverse proxy that spawns the binary
  from a `fastcgi/` parent.
- `server.ListenAndServe` precedence: `FCGI_LISTEN` env var → `core.IsFCGI()`
  → plain HTTP. The env-var path is what makes local Caddy dev work without
  faking the Hostsharing tree.
- `core.ServiceName()` precedence: `SERVICE_NAME` env var → executable
  basename with optional `.fcgi` suffix stripped.
- `server.ReadInConfig` precedence: `./<app>.conf` in cwd → per-domain
  config dir (Hostsharing) → `$XDG_CONFIG_HOME/<app>` →
  `$HOME/.config/<app>` → `$HOME/.<app>`.

## Testing

```sh
go test ./...
```

All packages are tested with table-driven subtests. The Hostsharing layout
itself is documented at <https://www.hostsharing.net/doc/managed-operations-platform/>.
