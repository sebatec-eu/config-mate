package core

import (
	"fmt"
	"os"
	"testing"
)

// serviceName is the internal helper. TestServiceName exercises both branches:
// env-var precedence and executable fallback (with .fcgi strip).
func TestServiceName(t *testing.T) {
	envTests := []struct {
		envValue string
		expected string
	}{
		{"my-service", "my-service"},
		{"api", "api"},
	}

	for _, tc := range envTests {
		t.Run("env_"+tc.envValue, func(t *testing.T) {
			t.Setenv(serviceNameEnvVar, tc.envValue)

			name, err := serviceName(func() (string, error) {
				return "/dummy/path", nil
			})
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if name != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, name)
			}
		})
	}

	fallbackTests := []struct {
		path     string
		expected string
		err      bool
	}{
		{"/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi-ssl/api.fcgi", "api", false},
		{"/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi-ssl/api", "api", false},
		{"/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi-ssl/hello-world.fcgi", "hello-world", false},
		{"/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi-ssl/hello-world", "hello-world", false},
		{"", "", true},
		{"/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi-ssl/.fcgi", "", true},
	}

	for _, tc := range fallbackTests {
		t.Run("fallback_"+tc.path, func(t *testing.T) {
			os.Unsetenv(serviceNameEnvVar)
			name, err := serviceName(func() (string, error) {
				if tc.path == "" {
					return "", fmt.Errorf("mock error")
				}
				return tc.path, nil
			})

			if tc.err {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if name != tc.expected {
					t.Errorf("Expected %q, got %q", tc.expected, name)
				}
			}
		})
	}
}

// TestXdgConfigHome pins the XDG precedence: XDG_CONFIG_HOME → $HOME/.config → "".
// This is env-only — no Hostsharing layout involved — so it belongs in core.
func TestXdgConfigHome(t *testing.T) {
	t.Run("XDG_CONFIG_HOME wins when set", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
		t.Setenv("HOME", "/home/me")

		if got := xdgConfigHome(); got != "/custom/xdg" {
			t.Errorf("got %q, want %q", got, "/custom/xdg")
		}
	})

	t.Run("falls back to $HOME/.config when XDG_CONFIG_HOME is empty", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("HOME", "/home/me")

		if got := xdgConfigHome(); got != "/home/me/.config" {
			t.Errorf("got %q, want %q", got, "/home/me/.config")
		}
	})

	t.Run("returns empty string when neither env var is set", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("HOME", "")

		if got := xdgConfigHome(); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}
