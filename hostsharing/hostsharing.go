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
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
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
// If neither condition is met, it starts a standard HTTP server on the default port.
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

	log.Printf("Server listening on port %s\n", defaultHttpPort)
	if err := http.ListenAndServe(":"+defaultHttpPort, handler); err != nil {
		return fmt.Errorf("http.ListenAndServe failed on port %s: %v", defaultHttpPort, err)
	}
	return nil
}

func base64StringToBytesHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf([]byte{}) {
			return data, nil
		}

		if result, err := base64.StdEncoding.DecodeString(data.(string)); err == nil {
			return result, nil
		}

		return data, nil
	}
}

// ReadInConfig reads and unmarshals configuration from a file into the provided value.
// It attempts to load configuration from a local file first, then falls back to
// searching in domain-specific and home directories. The function supports custom
// decode hooks for type conversion during unmarshaling.
//
// Parameters:
//   - rawVal: A pointer to a struct where the unmarshaled configuration will be stored.
//   - app_name: The application name used to construct config file paths. If empty,
//     it will be determined automatically via ServiceName(). Config files are expected
//     to be named as ".<app_name>.conf".
//   - fs: Optional variadic mapstructure.DecodeHookFunc functions for custom type
//     conversion during unmarshaling. If none are provided, default hooks are applied:
//     base64StringToBytesHookFunc(), StringToTimeDurationHookFunc(), and
//     StringToSliceHookFunc with "," delimiter.
//
// Returns:
//   - error: Returns nil on success, or an error describing what went wrong during
//     config file reading or unmarshaling.
//
// The function searches for config in the following order:
//  1. Local file: .<app_name>.conf
//  2. Domain-specific path: /home/pacs/xyz00/users/foobar/doms/example.com/etc/{app_name}
//  3. Home directory: $HOME/.<app_name>
func ReadInConfig(rawVal any, app_name string, fs ...mapstructure.DecodeHookFunc) error {
	if app_name == "" {
		a, err := ServiceName()
		if err != nil {
			return err
		}
		app_name = a
	}

	viper.SetConfigType("yaml")
	cfg, err := os.ReadFile(fmt.Sprintf(".%s.conf", app_name))
	if err != nil {
		domain, err := DomainByWorkingDir()
		if err != nil && err != ErrShortPath {
			panic(err)
		}
		if domain != nil {
			viper.AddConfigPath(fmt.Sprintf("%s/%s", domain.ConfigDir(), app_name))
		}
		viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", app_name))

		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("fatal error config file: %w", err)
		}
	} else {
		viper.ReadConfig(bytes.NewBuffer(cfg))
	}

	if len(fs) <= 0 {
		fs = append(fs,
			base64StringToBytesHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)
	}

	if err := viper.Unmarshal(&rawVal, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(fs...))); err != nil {
		return fmt.Errorf("cannot unmarshal config: %v", err)
	}

	return nil
}
