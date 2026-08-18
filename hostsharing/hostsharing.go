// Package hostsharing provides utilities for applications running in Hostsharing environments.
//
// It offers functionality for:
//   - Service name detection from environment variables or executable paths
//   - Parsing domain and user information from filesystem paths
//   - HTTP server configuration for both standard HTTP and FastCGI protocols
//   - Configuration file loading with support for domain-specific and user home directories
//   - Structured logging with request context tracking
//
// The package is designed to integrate with Hostsharing's directory structure where
// applications are organized as: /home/pacs/{pac}/users/{user}/doms/{domain}
package hostsharing

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path/filepath"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const defaultHttpPort = "9000"

// ErrNoFcgiEnvironment indicates that the FastCGI environment was not detected.
var ErrNoFcgiEnvironment = fmt.Errorf("no fcgi environment dedected")

// ListenAndServe starts an HTTP server using either FastCGI or standard HTTP,
// depending on the environment.
//
// It first checks for the FCGI_LISTEN environment variable. If set, it uses FastCGI with the specified address.
// Otherwise, it falls back to the existing IsFCGI() logic.
// If neither condition is met, it starts a standard HTTP server bound to the
// address resolved by [listenAddr]: ADDR (e.g. "127.0.0.1:9000") takes
// precedence; otherwise PORT (a bare port number, e.g. "8080") is used;
// otherwise the default port (":9000") is used.
//
// Example:
//
//	// Caddyfile configuration:
//	// {
//	//     auto_https off
//	//     http_port 1313
//	//     admin off
//	// }
//	// localhost:1313 {
//	//     root * public
//	//     file_server
//	//     reverse_proxy /api/* :9000 {
//	//         transport fastcgi
//	//     }
//	// }
//	// Set FCGI_LISTEN=:9000 in your environment to enable FastCGI mode.
//
// Example usage:
//
//	r := http.NewServeMux()
//	r.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
//	    fmt.Fprint(w, "Hello, FastCGI!")
//	})
//	if err := hostsharing.ListenAndServe(r); err != nil {
//	    log.Fatalf("Server failed: %v", err)
//	}
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

	if IsFCGI() {
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

func listenAddr() string {
	if addr := os.Getenv("ADDR"); addr != "" {
		return addr
	}
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":" + defaultHttpPort
}

// Base64StringToBytesHookFunc returns a mapstructure decode hook that turns
// string fields into []byte by trying each encoding in order. The first
// encoding that decodes the string wins; if none decode it, the string is
// returned unchanged. Pass one encoding to accept a single alphabet, several
// to accept a fallback chain (for example StdEncoding then URLEncoding).
// Padding rules are set on each encoding — use RawStdEncoding, RawURLEncoding,
// or Encoding.WithPadding for unpadded or custom-padded input.
func Base64StringToBytesHookFunc(encs ...*base64.Encoding) mapstructure.DecodeHookFuncType {
	return func(f, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf([]byte{}) {
			return data, nil
		}
		s, _ := data.(string)
		for _, enc := range encs {
			if result, err := enc.DecodeString(s); err == nil {
				return result, nil
			}
		}
		return data, nil
	}
}

// ReadInConfig reads and unmarshals configuration from a file into rawVal.
// The application name comes from app_name, falling back to [ServiceName]
// (SERVICE_NAME env, else executable) when empty.
//
// A missing config file is not an error: ReadInConfig returns nil and rawVal
// retains its zero values plus any viper.SetDefault calls made before this
// function (mapstructure `default:"..."` struct tags are NOT applied by viper).
//
// Search order:
//  1. .<app>.conf in the current working directory (loaded via ReadConfig).
//  2. <domain.ConfigDir>/<app> — e.g. /home/pacs/.../doms/example.com/etc/api
//     (resolved via [DomainByExecutable]; CONFIG_BASE_PATH is honored for local dev).
//  3. $XDG_CONFIG_HOME/<app>, falling back to $HOME/.config/<app>.
//  4. $HOME/.<app> (legacy).
//
// rawVal must be a pointer. fs optionally adds mapstructure decode hooks; when
// empty, defaults are: Base64StringToBytesHookFunc(Std, URL),
// mapstructure.StringToTimeDurationHookFunc, mapstructure.StringToSliceHookFunc(",").
func ReadInConfig(rawVal any, app_name string, fs ...mapstructure.DecodeHookFunc) error {
	if app_name == "" {
		name, err := ServiceName()
		if err != nil {
			return err
		}
		app_name = name
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(app_name)

	if cfg, err := os.ReadFile("." + app_name + ".conf"); err == nil {
		if err := v.ReadConfig(bytes.NewBuffer(cfg)); err != nil {
			return fmt.Errorf("cannot read local config: %w", err)
		}
	} else {
		if domain, err := DomainByExecutable(); err != nil && err != ErrShortPath {
			panic(err)
		} else if domain != nil {
			v.AddConfigPath(domain.ConfigDir())
		}
		if xdg := xdgConfigHome(); xdg != "" {
			v.AddConfigPath(xdg)
		}
		if home := os.Getenv("HOME"); home != "" {
			v.AddConfigPath(filepath.Join(home, "."+app_name))
		}

		if err := v.ReadInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return fmt.Errorf("cannot read config: %w", err)
		}
	}

	if len(fs) <= 0 {
		fs = append(fs,
			Base64StringToBytesHookFunc(base64.StdEncoding, base64.URLEncoding),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)
	}

	if err := v.Unmarshal(&rawVal, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(fs...))); err != nil {
		return fmt.Errorf("cannot unmarshal config: %v", err)
	}

	return nil
}

func xdgConfigHome() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config")
	}
	return ""
}
