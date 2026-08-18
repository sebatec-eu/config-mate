// Package server provides HTTP/FastCGI server plumbing for config-mate apps.
//
// It is environment-aware but not environment-bound:
//   - [ListenAndServe] honours FCGI_LISTEN first (so a Caddy reverse proxy
//     with `transport fastcgi` works in dev without faking the Hostsharing
//     tree), then [core.IsFCGI] for the real Hostsharing FastCGI case, then
//     plain HTTP for everything else.
//   - [RequestLogger] writes per-domain log files when running under
//     Hostsharing FastCGI and to stdout otherwise.
//   - [ReadInConfig] loads application configuration with sensible precedence
//     across cwd, per-domain config dir, XDG, and $HOME/.<app>.
//
// Everything in this package works on Hostsharing + dev AND on VM/Root + dev
// profiles. It depends on [core] (env-agnostic helpers) and [hostsharing]
// only for path-derived log file resolution.
package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"

	"github.com/sebatec-eu/config-mate/v2/core"
)

const defaultHttpPort = "9000"

// ListenAndServe starts an HTTP server using either FastCGI or standard HTTP,
// depending on the environment.
//
// Precedence:
//  1. FCGI_LISTEN env var → FastCGI on that address (lets Caddy
//     `reverse_proxy` + `transport fastcgi` work in dev where the binary is
//     NOT under a fastcgi/ parent directory).
//  2. core.IsFCGI() → FastCGI on stdin/stdout (Hostsharing's Apache alias
//     /fastcgi-bin/ invokes us from …/doms/<host>/fastcgi/, so the parent
//     dir's base name is "fastcgi" or "fastcgi-ssl").
//  3. Otherwise plain HTTP on the address resolved by [listenAddr]:
//     ADDR (e.g. "127.0.0.1:9000") → PORT (bare port, e.g. "8080") →
//     default ":9000".
func ListenAndServe(handler http.Handler) error {
	if addr := os.Getenv("FCGI_LISTEN"); addr != "" {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("net.Listen failed for FCGI_LISTEN=%s: %v", addr, err)
		}
		if err := fcgi.Serve(ln, handler); err != nil {
			return fmt.Errorf("fcgi.Serve failed on %s: %v", addr, err)
		}
		return nil
	}

	if core.IsFCGI() {
		if err := fcgi.Serve(nil, handler); err != nil {
			return fmt.Errorf("fcgi.Serve failed: %v", err)
		}
		return nil
	}

	addr := listenAddr()
	log.Printf("Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		return fmt.Errorf("http.ListenAndServe failed on %s: %v", addr, err)
	}
	return nil
}

// listenAddr resolves the listen address for the plain-HTTP branch.
// ADDR → PORT → default ":9000". ADDR accepts any host:port string passed
// through to http.ListenAndServe; PORT is a bare port number.
func listenAddr() string {
	if addr := os.Getenv("ADDR"); addr != "" {
		return addr
	}
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":" + defaultHttpPort
}
