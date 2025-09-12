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

func ReadInConfig(rawVal any, app_name string, fs ...mapstructure.DecodeHookFunc) error {
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
