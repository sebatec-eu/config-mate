package hostsharing

import (
	"fmt"
	"os"
	"testing"
)

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
			oldEnv := os.Getenv(serviceNameEnvVar)
			t.Setenv(serviceNameEnvVar, tc.envValue)
			defer os.Setenv(serviceNameEnvVar, oldEnv)

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
