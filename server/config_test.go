package server

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ReadInConfig does not honour mapstructure `default:` tags; missing configs
// leave Foo at its zero value, which is what most subtests assert.

func TestReadInConfig(t *testing.T) {
	const (
		appName    = "myapp"
		baseYAML   = "foo: from-config\n"
		brokenYAML = "foo: [unterminated\n: :"
	)

	tests := []struct {
		name    string
		env     map[string]string
		files   map[string]string // relative to $HOME
		chdir   bool
		wantErr bool
		wantFoo string
	}{
		{"no file", map[string]string{"SERVICE_NAME": appName}, nil, false, false, ""},
		{"XDG explicit", map[string]string{
			"SERVICE_NAME": appName, "XDG_CONFIG_HOME": "XDG_PLACEHOLDER/xdg", "CONFIG_BASE_PATH": "",
		}, map[string]string{"xdg/myapp.yaml": baseYAML}, false, false, "from-config"},
		{"propagates parse errors", map[string]string{
			"SERVICE_NAME": appName, "XDG_CONFIG_HOME": "", "CONFIG_BASE_PATH": "",
		}, map[string]string{".config/myapp.yaml": brokenYAML}, false, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			for k, v := range tt.env {
				if v == "XDG_PLACEHOLDER/xdg" {
					v = filepath.Join(home, "xdg")
				}
				t.Setenv(k, v)
			}
			t.Setenv("HOME", home)
			for rel, content := range tt.files {
				abs := filepath.Join(home, rel)
				if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if tt.chdir {
				t.Chdir(home)
			}

			var cfg struct{ Foo string }
			err := ReadInConfig(&cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got nil (cfg=%+v)", cfg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Foo != tt.wantFoo {
				t.Fatalf("Foo: want %q, got %q", tt.wantFoo, cfg.Foo)
			}
		})
	}
}

// Subdir <xdg>/<app>/<app>.yaml must win over flat <xdg>/<app>.yaml.
func TestReadInConfig_XDGSubdirWinsOverFlat(t *testing.T) {
	home := t.TempDir()
	xdg := filepath.Join(home, "xdg")
	app := "myapp"
	t.Setenv("SERVICE_NAME", app)
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv("CONFIG_BASE_PATH", "")
	t.Setenv("HOME", home)

	for _, p := range []string{xdg, filepath.Join(xdg, app)} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite(t, filepath.Join(xdg, app+".yaml"), "foo: from-flat\n")
	mustWrite(t, filepath.Join(xdg, app, app+".yaml"), "foo: from-subdir\n")

	var cfg struct{ Foo string }
	if err := ReadInConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Foo != "from-subdir" {
		t.Fatalf("want from-subdir, got %q", cfg.Foo)
	}
}

// $HOME/.config fallback when XDG_CONFIG_HOME is unset.
func TestReadInConfig_XDGFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("SERVICE_NAME", "myapp")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("CONFIG_BASE_PATH", "")
	t.Setenv("HOME", home)

	mustWrite(t, filepath.Join(home, ".config", "myapp.yaml"), "foo: from-config\n")

	var cfg struct{ Foo string }
	if err := ReadInConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Foo != "from-config" {
		t.Fatalf("want from-config, got %q", cfg.Foo)
	}
}

// Legacy $HOME/.<app> directory.
func TestReadInConfig_LegacyHomeDot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("SERVICE_NAME", "myapp")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("CONFIG_BASE_PATH", "")
	t.Setenv("HOME", home)

	mustWrite(t, filepath.Join(home, ".myapp", "myapp.yaml"), "foo: from-config\n")

	var cfg struct{ Foo string }
	if err := ReadInConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Foo != "from-config" {
		t.Fatalf("want from-config, got %q", cfg.Foo)
	}
}

// Non-ErrShortPath hostsharing failure must log and fall through to XDG.
func TestReadInConfig_HostsharingErrorIsLogged(t *testing.T) {
	withStubbedHostsharing(t, func() (string, error) {
		return "", fmt.Errorf("synthetic hostsharing failure")
	})

	home := t.TempDir()
	xdg := filepath.Join(home, "xdg")
	t.Setenv("SERVICE_NAME", "myapp")
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv("CONFIG_BASE_PATH", "")
	t.Setenv("HOME", home)

	mustWrite(t, filepath.Join(xdg, "myapp.yaml"), "foo: from-xdg\n")

	var cfg struct{ Foo string }
	if err := ReadInConfig(&cfg); err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if cfg.Foo != "from-xdg" {
		t.Fatalf("want from-xdg, got %q", cfg.Foo)
	}
}

// ErrShortPath (no PAC layout) is the normal non-Hostsharing case and must
// stay silent — no "PAC detection failed" log noise on plain HTTP boots.
func TestReadInConfig_HostsharingErrShortPathIsSilent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("SERVICE_NAME", "myapp")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("CONFIG_BASE_PATH", "")
	t.Setenv("HOME", home)

	var buf bytes.Buffer
	origLog := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(origLog) })

	var cfg struct{ Foo string }
	if err := ReadInConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "PAC detection failed") {
		t.Fatalf("ErrShortPath must be silent, got: %s", buf.String())
	}
}

// --- helpers ---

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func withStubbedHostsharing(t *testing.T, stub func() (string, error)) {
	t.Helper()
	orig := hostsharingConfigDir
	hostsharingConfigDir = stub
	t.Cleanup(func() { hostsharingConfigDir = orig })

	var buf bytes.Buffer
	origLog := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(origLog) })
}
