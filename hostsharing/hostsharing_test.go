package hostsharing

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadInConfig is a table-driven test of [ReadInConfig]'s file-search
// behaviour and missing-config tolerance.
//
// viper does not honour mapstructure `default:"..."` struct tags (it leaves
// DecoderConfig.Metadata as nil), so a missing config leaves Foo at its
// zero value — that is the "defaults preserved" behaviour under test.
func TestReadInConfig(t *testing.T) {
	const appName = "myapp"
	const baseYAML = "foo: from-config\n"
	const brokenYAML = "foo: [unterminated\n: :"

	tests := []struct {
		name    string
		env     map[string]string
		files   map[string]string // paths relative to $HOME
		chdir   bool              // chdir into $HOME before calling ReadInConfig
		wantErr bool
		wantFoo string
	}{
		{
			name:    "no file returns nil and leaves zero values",
			env:     map[string]string{"SERVICE_NAME": appName},
			wantErr: false,
			wantFoo: "",
		},
		{
			name: "loads from explicit XDG_CONFIG_HOME",
			env: map[string]string{
				"SERVICE_NAME":     appName,
				"XDG_CONFIG_HOME":  "XDG_PLACEHOLDER/xdg", // resolved to t.TempDir() per subtest
				"CONFIG_BASE_PATH": "",
			},
			files: map[string]string{
				"xdg/myapp.yaml": baseYAML,
			},
			wantErr: false,
			wantFoo: "from-config",
		},
		{
			name: "local .<app>.conf in CWD is loaded",
			env: map[string]string{
				"SERVICE_NAME":    appName,
				"XDG_CONFIG_HOME": "",
			},
			files: map[string]string{
				".myapp.conf": "foo: from-local\n",
			},
			chdir:   true,
			wantErr: false,
			wantFoo: "from-local",
		},
		{
			name: "propagates non-not-found errors",
			env: map[string]string{
				"SERVICE_NAME":     appName,
				"XDG_CONFIG_HOME":  "",
				"CONFIG_BASE_PATH": "",
			},
			files: map[string]string{
				".config/myapp.yaml": brokenYAML,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()

			// Resolve XDG_CONFIG_HOME placeholder to the per-subject temp dir.
			for k, v := range tt.env {
				if v == "XDG_PLACEHOLDER/xdg" {
					v = filepath.Join(home, "xdg")
				}
				t.Setenv(k, v)
			}
			t.Setenv("HOME", home)

			for relPath, content := range tt.files {
				abs := filepath.Join(home, relPath)
				if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
					t.Fatalf("write: %v", err)
				}
			}

			if tt.chdir {
				t.Chdir(home)
			}

			var cfg struct {
				Foo string
			}
			err := ReadInConfig(&cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (cfg=%+v)", cfg)
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

// TestReadInConfig_XDGFallback verifies that, when XDG_CONFIG_HOME is unset,
// ReadInConfig falls back to $HOME/.config and loads <app>.yaml from there.
func TestReadInConfig_XDGFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("SERVICE_NAME", "myapp")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("CONFIG_BASE_PATH", "")
	t.Setenv("HOME", home)

	dir := filepath.Join(home, ".config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "myapp.yaml"), []byte("foo: from-config\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var cfg struct {
		Foo string
	}
	if err := ReadInConfig(&cfg); err != nil {
		t.Fatalf("ReadInConfig: %v", err)
	}
	if cfg.Foo != "from-config" {
		t.Fatalf("Foo: want %q, got %q", "from-config", cfg.Foo)
	}
}

// TestReadInConfig_LegacyHomeDot verifies that the legacy $HOME/.<app>
// directory is still searched.
func TestReadInConfig_LegacyHomeDot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("SERVICE_NAME", "myapp")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("CONFIG_BASE_PATH", "")
	t.Setenv("HOME", home)

	dir := filepath.Join(home, ".myapp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "myapp.yaml"), []byte("foo: from-config\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var cfg struct {
		Foo string
	}
	if err := ReadInConfig(&cfg); err != nil {
		t.Fatalf("ReadInConfig: %v", err)
	}
	if cfg.Foo != "from-config" {
		t.Fatalf("Foo: want %q, got %q", "from-config", cfg.Foo)
	}
}

// TestListenAndServeAddr is a table-driven test of the listenAddr() pure
// function used by [ListenAndServe]'s HTTP branch.
//
// Each subtest explicitly sets ADDR and PORT with t.Setenv (never iterating
// over a map) so the precedence rule is unambiguous. t.Setenv also seeds
// both vars to "" where unset, so the case does not depend on the outer
// process environment.
func TestListenAndServeAddr(t *testing.T) {
	tests := []struct {
		name string
		addr string
		port string
		want string
	}{
		{
			name: "default: both unset",
			want: ":" + defaultHttpPort,
		},
		{
			name: "PORT only",
			port: "8080",
			want: ":8080",
		},
		{
			name: "ADDR only",
			addr: "127.0.0.1:9000",
			want: "127.0.0.1:9000",
		},
		{
			name: "ADDR overrides PORT",
			addr: "127.0.0.1:9000",
			port: "8080",
			want: "127.0.0.1:9000",
		},
		{
			name: "ADDR with explicit IPv4 wildcard",
			addr: "0.0.0.0:9090",
			want: "0.0.0.0:9090",
		},
		{
			name: "empty strings fall through to default",
			addr: "",
			port: "",
			want: ":" + defaultHttpPort,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Explicit t.Setenv per var — never iterate a map. Sets to ""
			// when the case does not specify a value, so the test is
			// independent of the outer process env.
			t.Setenv("ADDR", tc.addr)
			t.Setenv("PORT", tc.port)

			if got := listenAddr(); got != tc.want {
				t.Errorf("listenAddr() = %q, want %q", got, tc.want)
			}
		})
	}
}
