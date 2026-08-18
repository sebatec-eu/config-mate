package hostsharing

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mitchellh/mapstructure"
)

func TestBase64StringToBytesHookFunc(t *testing.T) {
	std := "AAECAwQFBgcICQoLDA0ODw==" // 16 bytes: 0x00..0x0f
	want := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

	// URL-safe fixture: same plaintext, alphabet uses '-' and '_'.
	url := base64.URLEncoding.EncodeToString(want)

	t.Run("StdEncoding round-trips known fixture", func(t *testing.T) {
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), std)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, ok := got.([]byte)
		if !ok {
			t.Fatalf("expected []byte, got %T", got)
		}
		if !bytes.Equal(b, want) {
			t.Errorf("expected %x, got %x", want, b)
		}
	})

	t.Run("URLEncoding round-trips URL-safe fixture", func(t *testing.T) {
		got, err := runHook(Base64StringToBytesHookFunc(base64.URLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), url)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, ok := got.([]byte)
		if !ok {
			t.Fatalf("expected []byte, got %T", got)
		}
		if !bytes.Equal(b, want) {
			t.Errorf("expected %x, got %x", want, b)
		}
	})

	t.Run("Std then URL: Std-only input decodes via Std", func(t *testing.T) {
		// Std has '+' and '/'; URL rejects them. If Std decodes first, we get the bytes.
		stdPlus := base64.StdEncoding.EncodeToString([]byte{0xfb, 0xff, 0xfe}) // contains '/'
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding, base64.URLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), stdPlus)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(got.([]byte), []byte{0xfb, 0xff, 0xfe}) {
			t.Errorf("expected Std decoding to win, got %x", got)
		}
	})

	t.Run("Std then URL: URL-only chars fall through to URL", func(t *testing.T) {
		// Force a '-' into the encoding by using RawURLEncoding on a byte slice that
		// happens to produce one. Simplest: build a URL-only string directly.
		urlOnly := "-__-" // valid only in URL alphabet
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding, base64.URLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), urlOnly)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == urlOnly {
			t.Errorf("expected URL decoding to succeed; input was returned unchanged")
		}
	})

	t.Run("First match wins when input is valid in both alphabets", func(t *testing.T) {
		// 'AAA=' is valid padded input in both Std and URL alphabets. Both
		// decode it to the same byte. We verify the hook returns successfully;
		// the bytes are identical for both alphabets so we cannot assert which
		// one ran — only that the first-match-wins semantics did not error.
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding, base64.URLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), "AAA=")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(got.([]byte), []byte{0x00, 0x00}) {
			t.Errorf("expected [0x00 0x00], got %x", got)
		}
	})

	t.Run("RawURLEncoding accepts unpadded URL input", func(t *testing.T) {
		raw := base64.RawURLEncoding.EncodeToString(want)
		got, err := runHook(Base64StringToBytesHookFunc(base64.RawURLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(got.([]byte), want) {
			t.Errorf("expected %x, got %x", want, got)
		}
	})

	t.Run("URLEncoding rejects unpadded input", func(t *testing.T) {
		raw := base64.RawURLEncoding.EncodeToString(want)
		got, err := runHook(Base64StringToBytesHookFunc(base64.URLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != raw {
			t.Errorf("expected input to pass through unchanged when padding is wrong, got %v", got)
		}
	})

	t.Run("No encodings: input passes through", func(t *testing.T) {
		got, err := runHook(Base64StringToBytesHookFunc(), reflect.TypeOf(""), reflect.TypeOf([]byte{}), std)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != std {
			t.Errorf("expected pass-through, got %v", got)
		}
	})

	t.Run("Non-string source passes through", func(t *testing.T) {
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding), reflect.TypeOf(int(0)), reflect.TypeOf([]byte{}), 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 42 {
			t.Errorf("expected pass-through, got %v", got)
		}
	})

	t.Run("Non-[]byte target passes through", func(t *testing.T) {
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding), reflect.TypeOf(""), reflect.TypeOf(int(0)), std)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != std {
			t.Errorf("expected pass-through, got %v", got)
		}
	})

	t.Run("All encodings reject: input passes through", func(t *testing.T) {
		bad := "###not base64###"
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding, base64.URLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), bad)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != bad {
			t.Errorf("expected pass-through on total mismatch, got %v", got)
		}
	})

	t.Run("ComposeDecodeHookFunc accepts the hook", func(t *testing.T) {
		// ComposeDecodeHookFunc is the integration point used by ReadInConfig.
		// If the hook's signature stops matching mapstructure's expectations,
		// this composition step will fail to compile.
		_ = mapstructure.ComposeDecodeHookFunc(
			Base64StringToBytesHookFunc(base64.StdEncoding, base64.URLEncoding),
		)
	})
}

// runHook invokes a DecodeHookFunc with explicit from/to reflect.Types, mirroring
// how mapstructure.ComposeDecodeHookFunc calls them.
func runHook(f mapstructure.DecodeHookFuncType, from, to reflect.Type, data any) (any, error) {
	return f(from, to, data)
}

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

			// Set env in a fixed, explicit order — map iteration is non-deterministic,
			// and order matters when one env var resolves to a path under $HOME.
			setEnvIfPresent := func(k string) {
				if v, ok := tt.env[k]; ok {
					if v == "XDG_PLACEHOLDER/xdg" {
						v = filepath.Join(home, "xdg")
					}
					t.Setenv(k, v)
				} else {
					t.Setenv(k, "")
				}
			}
			setEnvIfPresent("SERVICE_NAME")
			setEnvIfPresent("XDG_CONFIG_HOME")
			setEnvIfPresent("CONFIG_BASE_PATH")
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
			err := ReadInConfig(&cfg, "")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (cfg=%v)", cfg)
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
	if err := ReadInConfig(&cfg, ""); err != nil {
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
	if err := ReadInConfig(&cfg, ""); err != nil {
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
