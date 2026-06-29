package hostsharing

import (
	"testing"
)

func TestParseUser(t *testing.T) {
	for _, tc := range []struct {
		path     string
		expected string
	}{
		{"/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi-ssl/api.fcgi", "xyz00-foobar"},
		{"/home/pacs/xyz00/users/foobar/doms/example.com", "xyz00-foobar"},
		{"/home/pacs/xyz00/users/foobar/", "xyz00-foobar"},
		{"/home/pacs/xyz00/users/foobar", "xyz00-foobar"},
		{"/home/pacs/xyz00", "xyz00"},
		{"/home/pacs/xyz00/", "xyz00"},
		{"/home/pacs/xyz00/users", "xyz00"},
	} {
		u, err := ParseUser(tc.path)
		if err != nil {
			t.Errorf("Got error: %s", err)
		}

		got, err := u.User()
		if got == "" {
			// PAC-only paths: a user sub-account is genuinely missing.
			if err != ErrNoUser {
				t.Errorf("Expected ErrNoUser for PAC-only path %q but got %v", tc.path, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if got != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, got)
		}
	}

	for _, tc := range []struct {
		path     string
		expected error
	}{
		{"", ErrShortPath},
		{"/home/pacs", ErrShortPath},
		{"/home/pacs/", ErrShortPath},
	} {
		u, err := ParseUser(tc.path)
		if err == nil {
			t.Error("Expected error but got nil")
		}

		if u != nil {
			t.Error("Got value instead of nil")
		}

		if err != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, err)
		}
	}
}

func TestParseDomainAnchor(t *testing.T) {
	for _, tc := range []struct {
		path           string
		expectedDomain string
	}{
		{"/srv/doms/example.com/fastcgi-ssl/api.fcgi", "example.com"},
		{"/srv/doms/example.com/fastcgi/api.fcgi", "example.com"},
		{"/srv/doms/example.com/fastcgi-ssl", "example.com"},
		{"/srv/doms/example.com", "example.com"},
		{"/srv/doms/example.org/etc/config.yaml", "example.org"},
	} {
		d, err := parseDomainFromBase(tc.path)
		if err != nil {
			t.Errorf("Expected no error for %q but got %v", tc.path, err)
			continue
		}
		if d.HasPAC() {
			t.Errorf("Expected no PAC for dev path %q but got pac=%q", tc.path, mustPAC(t, d))
		}
		if got := d.Domain(); got != tc.expectedDomain {
			t.Errorf("Expected domain %q for %q but got %q", tc.expectedDomain, tc.path, got)
		}
	}

	for _, tc := range []struct {
		path     string
		expected error
	}{
		{"", ErrShortPath},
		{"/srv", ErrShortPath},
		{"/srv/example.com", ErrShortPath}, // no "doms" anchor
	} {
		_, err := parseDomainFromBase(tc.path)
		if err != tc.expected {
			t.Errorf("Expected %v for %q but got %v", tc.expected, tc.path, err)
		}
	}
}

func mustPAC(t *testing.T, d *domain) string {
	t.Helper()
	pac, err := d.PAC()
	if err != nil {
		t.Fatalf("unexpected PAC error: %v", err)
	}
	return pac
}

func TestParseDomain(t *testing.T) {
	for _, tc := range []struct {
		path           string
		expectedUser   string
		expectedDomain string
	}{
		{"/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi-ssl/api.fcgi", "xyz00-foobar", "example.com"},
		{"/home/pacs/xyz00/users/foobar/doms/example.com", "xyz00-foobar", "example.com"},
	} {
		u, err := ParseDomain(tc.path)
		if err != nil {
			t.Errorf("Got error: %s", err)
		}

		got, err := u.User()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if got != tc.expectedUser {
			t.Errorf("Expected %s but got %s", tc.expectedUser, got)
		}

		if got := u.Domain(); got != tc.expectedDomain {
			t.Errorf("Expected %s but got %s", tc.expectedUser, got)
		}
	}

	for _, tc := range []struct {
		path     string
		expected error
	}{
		{"", ErrShortPath},
		{"/home/pacs", ErrShortPath},
		{"/home/pacs/", ErrShortPath},
		{"/home/pacs/xyz00/users/foobar/doms/", ErrShortPath},
		{"/home/pacs/xyz00/users/foobar/doms", ErrShortPath},
		{"/home/pacs/xyz00/users/", ErrShortPath},
		{"/home/pacs/xyz00/users", ErrShortPath},
		{"/home/pacs/xyz00", ErrShortPath},
	} {
		u, err := ParseDomain(tc.path)
		if err == nil {
			t.Error("Expected error but got nil")
		}

		if u != nil {
			t.Error("Got value instead of nil")
		}

		if err != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, err)
		}
	}
}

func TestUserUser(t *testing.T) {
	for _, tc := range []struct {
		user
		expected string
	}{
		{user{"xyz00", nil}, ""},
		{user{"xyz00", &[]string{"example"}[0]}, "xyz00-example"},
		{user{"xyz00", &[]string{"www.example.com"}[0]}, "xyz00-www.example.com"},
	} {
		got, err := tc.User()
		if got == "" {
			// PAC-only paths: now ErrNoUser is the documented return.
			if err != ErrNoUser {
				t.Errorf("Expected ErrNoUser for PAC-only user but got %v", err)
			}
			continue
		}
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if got != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, got)
		}
	}
}

func TestUserPAC(t *testing.T) {
	for _, tc := range []struct {
		user
		expected string
	}{
		{user{"xyz00", nil}, "xyz00"},
		{user{"xyz00", &[]string{"example"}[0]}, "xyz00"},
		{user{"xyz00", &[]string{"www.example.com"}[0]}, "xyz00"},
	} {
		got, err := tc.PAC()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if got != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, got)
		}
	}
}

func TestUserPACError(t *testing.T) {
	u := &user{}
	got, err := u.PAC()
	if err != ErrNoPAC {
		t.Errorf("Expected ErrNoPAC but got %v", err)
	}
	if got != "" {
		t.Errorf("Expected empty string but got %q", got)
	}
}

func TestUserUserError(t *testing.T) {
	if _, err := (&user{}).User(); err != ErrNoPAC {
		t.Errorf("Expected ErrNoPAC for empty user but got %v", err)
	}
	if _, err := (&user{pac: "xyz00"}).User(); err != ErrNoUser {
		t.Errorf("Expected ErrNoUser for PAC-only user but got %v", err)
	}
}

func TestUserHasPAC(t *testing.T) {
	if got := (&user{}).HasPAC(); got != false {
		t.Errorf("Expected false for empty user but got %v", got)
	}
	if got := (&user{pac: "xyz00"}).HasPAC(); got != true {
		t.Errorf("Expected true for PAC-only user but got %v", got)
	}
	u := &[]string{"example"}[0]
	if got := (&user{pac: "xyz00", user: u}).HasPAC(); got != true {
		t.Errorf("Expected true for full user but got %v", got)
	}
}

func TestUserHasUser(t *testing.T) {
	if got := (&user{}).HasUser(); got != false {
		t.Errorf("Expected false for empty user but got %v", got)
	}
	if got := (&user{pac: "xyz00"}).HasUser(); got != false {
		t.Errorf("Expected false for PAC-only user but got %v", got)
	}
	u := &[]string{"example"}[0]
	if got := (&user{pac: "xyz00", user: u}).HasUser(); got != true {
		t.Errorf("Expected true for full user but got %v", got)
	}
}

func TestUserHome(t *testing.T) {
	for _, tc := range []struct {
		user
		expected string
	}{
		{user{"xyz00", nil}, "/home/pacs/xyz00"},
		{user{"xyz00", &[]string{"example"}[0]}, "/home/pacs/xyz00/users/example"},
		{user{"xyz00", &[]string{"www.example.com"}[0]}, "/home/pacs/xyz00/users/www.example.com"},
	} {
		if got := tc.Home(); got != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, got)
		}
	}
}

func TestDomainHome(t *testing.T) {
	for _, tc := range []struct {
		domain
		expected string
	}{
		{domain{user: user{"xyz00", nil}, domain: "example.com"}, "/home/pacs/xyz00"},
		{domain{user: user{"xyz00", &[]string{"example"}[0]}, domain: "example.com"}, "/home/pacs/xyz00/users/example"},
		{domain{user: user{"xyz00", &[]string{"www.example.com"}[0]}, domain: "example.com"}, "/home/pacs/xyz00/users/www.example.com"},
	} {
		if got := tc.Home(); got != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, got)
		}
	}
}

func TestDomainConfigDir(t *testing.T) {
	for _, tc := range []struct {
		domain
		expected string
	}{
		{domain{user: user{"xyz00", nil}, domain: "example.com"}, "/home/pacs/xyz00/doms/example.com/etc"},
		{domain{user: user{"xyz00", &[]string{"example"}[0]}, domain: "example.com"}, "/home/pacs/xyz00/users/example/doms/example.com/etc"},
		{domain{user: user{"xyz00", &[]string{"www.example.com"}[0]}, domain: "example.com"}, "/home/pacs/xyz00/users/www.example.com/doms/example.com/etc"},
	} {
		if got := tc.ConfigDir(); got != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, got)
		}
	}
}

func TestUserConfigDir(t *testing.T) {
	for _, tc := range []struct {
		user
		expected string
	}{
		{user{"xyz00", nil}, "/home/pacs/xyz00/etc"},
		{user{"xyz00", &[]string{"example"}[0]}, "/home/pacs/xyz00/users/example/etc"},
		{user{"xyz00", &[]string{"www.example.com"}[0]}, "/home/pacs/xyz00/users/www.example.com/etc"},
	} {
		if got := tc.ConfigDir(); got != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, got)
		}
	}
}

func TestUserLogDir(t *testing.T) {
	for _, tc := range []struct {
		user
		expected string
	}{
		{user{"xyz00", nil}, "/home/pacs/xyz00/var"},
		{user{"xyz00", &[]string{"example"}[0]}, "/home/pacs/xyz00/users/example/var"},
		{user{"xyz00", &[]string{"www.example.com"}[0]}, "/home/pacs/xyz00/users/www.example.com/var"},
	} {
		if got := tc.LogDir(); got != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, got)
		}
	}
}

func TestDomainLogDir(t *testing.T) {
	for _, tc := range []struct {
		domain
		expected string
	}{
		{domain{user: user{"xyz00", nil}, domain: "example.com"}, "/home/pacs/xyz00/doms/example.com/var"},
		{domain{user: user{"xyz00", &[]string{"example"}[0]}, domain: "example.com"}, "/home/pacs/xyz00/users/example/doms/example.com/var"},
		{domain{user: user{"xyz00", &[]string{"www.example.com"}[0]}, domain: "example.com"}, "/home/pacs/xyz00/users/www.example.com/doms/example.com/var"},
	} {
		if got := tc.LogDir(); got != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, got)
		}
	}
}

func TestDomainDomain(t *testing.T) {
	for _, tc := range []struct {
		domain
		expected string
	}{
		{domain{user: user{"xyz00", nil}, domain: "example.com"}, "example.com"},
		{domain{user: user{"xyz00", &[]string{"example"}[0]}, domain: "example.com"}, "example.com"},
		{domain{user: user{"xyz00", &[]string{"www.example.com"}[0]}, domain: "example.org"}, "example.org"},
	} {
		if got := tc.Domain(); got != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, got)
		}
	}
}

func TestDomainDomsDir(t *testing.T) {
	for _, tc := range []struct {
		domain
		expected string
	}{
		{domain{user: user{"xyz00", nil}, domain: "example.com"}, "/home/pacs/xyz00/doms/example.com"},
		{domain{user: user{"xyz00", &[]string{"example"}[0]}, domain: "example.com"}, "/home/pacs/xyz00/users/example/doms/example.com"},
		{domain{user: user{"xyz00", &[]string{"www.example.com"}[0]}, domain: "example.org"}, "/home/pacs/xyz00/users/www.example.com/doms/example.org"},
	} {
		if got := tc.DomsDir(); got != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, got)
		}
	}
}

func TestDomainDirsDev(t *testing.T) {
	// When PAC is missing the dirs are anchored at the parsed base path,
	// relative to the parent directory of the doms/{host} segment.
	for _, tc := range []struct {
		path        string
		wantDomsDir string
		wantCfgDir  string
		wantLogDir  string
		wantDataDir string
	}{
		{
			path:        "/srv/doms/example.com/fastcgi-ssl/api.fcgi",
			wantDomsDir: "/srv/doms/example.com",
			wantCfgDir:  "/srv/doms/example.com/etc",
			wantLogDir:  "/srv/doms/example.com/var",
			wantDataDir: "/srv/doms/example.com/data",
		},
		{
			path:        "/srv/doms/example.com",
			wantDomsDir: "/srv/doms/example.com",
			wantCfgDir:  "/srv/doms/example.com/etc",
			wantLogDir:  "/srv/doms/example.com/var",
			wantDataDir: "/srv/doms/example.com/data",
		},
	} {
		d, err := parseDomainFromBase(tc.path)
		if err != nil {
			t.Fatalf("parseDomainFromBase(%q): %v", tc.path, err)
		}
		if d.HasPAC() {
			t.Fatalf("expected no PAC for dev path %q", tc.path)
		}
		if got := d.DomsDir(); got != tc.wantDomsDir {
			t.Errorf("DomsDir for %q: expected %q but got %q", tc.path, tc.wantDomsDir, got)
		}
		if got := d.ConfigDir(); got != tc.wantCfgDir {
			t.Errorf("ConfigDir for %q: expected %q but got %q", tc.path, tc.wantCfgDir, got)
		}
		if got := d.LogDir(); got != tc.wantLogDir {
			t.Errorf("LogDir for %q: expected %q but got %q", tc.path, tc.wantLogDir, got)
		}
		if got := d.DataDir(); got != tc.wantDataDir {
			t.Errorf("DataDir for %q: expected %q but got %q", tc.path, tc.wantDataDir, got)
		}
	}
}

func TestIsFCGI(t *testing.T) {
	for _, tc := range []struct {
		path     string
		expected bool
	}{
		{"/", false},
		{"/home/pacs/xyz00/users/example/doms/example.com/fastcgi-ssl/api.fcgi", true},
		{"/home/pacs/xyz00/users/example/doms/example.com/fastcgi-ssl/foobar.fcgi", true},
		{"/home/pacs/xyz00/users/example/doms/example.com/fastcgi/foobar.fcgi", true},
		{"/home/pacs/xyz00/users/example/doms/example.com/cgi/foobar.fcgi", false},
	} {
		if got := isFCGI(func() (string, error) { return tc.path, nil }); got != tc.expected {
			t.Errorf("Expected %v for %v but got %v", tc.expected, tc.path, got)
		}
	}
}

func TestDomainByExecutable(t *testing.T) {
	noEnv := func(string) string { return "" }

	// Test successful cases
	for _, tc := range []struct {
		name           string
		envLookup      func(string) string
		getExecutable  func() (string, error)
		expectedUser   string
		expectedDomain string
	}{
		{
			name:      "valid executable in fastcgi-ssl directory",
			envLookup: noEnv,
			getExecutable: func() (string, error) {
				return "/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi-ssl/api.fcgi", nil
			},
			expectedUser:   "xyz00-foobar",
			expectedDomain: "example.com",
		},
		{
			name:      "valid executable in fastcgi directory",
			envLookup: noEnv,
			getExecutable: func() (string, error) {
				return "/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi/api.fcgi", nil
			},
			expectedUser:   "xyz00-foobar",
			expectedDomain: "example.com",
		},
		{
			name:      "valid executable with trailing slash",
			envLookup: noEnv,
			getExecutable: func() (string, error) {
				return "/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi-ssl/api.fcgi/", nil
			},
			expectedUser:   "xyz00-foobar",
			expectedDomain: "example.com",
		},
		{
			name:           "valid executable with minimum path",
			envLookup:      noEnv,
			getExecutable:  func() (string, error) { return "/home/pacs/abc/users/def/doms/test.org/fastcgi-ssl/api.fcgi", nil },
			expectedUser:   "abc-def",
			expectedDomain: "test.org",
		},
		{
			name:           "CONFIG_BASE_PATH set to valid fastcgi-ssl path",
			envLookup:      func(string) string { return "/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi-ssl/api.fcgi" },
			getExecutable:  func() (string, error) { return "/tmp/go-build/whatever/exe/main", nil },
			expectedUser:   "xyz00-foobar",
			expectedDomain: "example.com",
		},
		{
			name: "CONFIG_BASE_PATH wins over executable",
			envLookup: func(string) string {
				return "/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi-ssl/api.fcgi"
			},
			getExecutable: func() (string, error) {
				return "/home/pacs/other/users/different/doms/other.org/fastcgi-ssl/api.fcgi", nil
			},
			expectedUser:   "xyz00-foobar",
			expectedDomain: "example.com",
		},
		{
			name:           "CONFIG_BASE_PATH set to bare doms path",
			envLookup:      func(string) string { return "/home/pacs/xyz00/users/foobar/doms/example.com" },
			getExecutable:  func() (string, error) { return "/tmp/go-build/whatever/exe/main", nil },
			expectedUser:   "xyz00-foobar",
			expectedDomain: "example.com",
		},
		{
			// CONFIG_BASE_PATH is too shallow even after filepath.Dir falls back
			// to "/". The function then falls through to the executable branch,
			// which succeeds — dev-opt-in behavior.
			name:      "CONFIG_BASE_PATH too shallow falls through to executable",
			envLookup: func(string) string { return "/api.fcgi" },
			getExecutable: func() (string, error) {
				return "/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi-ssl/api.fcgi", nil
			},
			expectedUser:   "xyz00-foobar",
			expectedDomain: "example.com",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := domainByExecutable(tc.envLookup, tc.getExecutable)
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if d == nil {
				t.Fatal("Expected domain but got nil")
			}

			got, err := d.User()
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if got != tc.expectedUser {
				t.Errorf("Expected user %s but got %s", tc.expectedUser, got)
			}

			if got := d.Domain(); got != tc.expectedDomain {
				t.Errorf("Expected domain %s but got %s", tc.expectedDomain, got)
			}
		})
	}

	// Test error cases
	for _, tc := range []struct {
		name          string
		envLookup     func(string) string
		getExecutable func() (string, error)
		expectedError error
	}{
		{
			name:          "getExecutable returns error",
			envLookup:     noEnv,
			getExecutable: func() (string, error) { return "", ErrShortPath },
			expectedError: ErrShortPath,
		},
		{
			name:          "getExecutable returns empty string",
			envLookup:     noEnv,
			getExecutable: func() (string, error) { return "", nil },
			expectedError: ErrShortPath,
		},
		{
			name:          "executable in pac directory (too shallow)",
			envLookup:     noEnv,
			getExecutable: func() (string, error) { return "/home/pacs/xyz00/api.fcgi", nil },
			expectedError: ErrShortPath,
		},
		{
			name:          "executable in users directory (too shallow)",
			envLookup:     noEnv,
			getExecutable: func() (string, error) { return "/home/pacs/xyz00/users/api.fcgi", nil },
			expectedError: ErrShortPath,
		},
		{
			name:          "executable in root (too shallow)",
			envLookup:     noEnv,
			getExecutable: func() (string, error) { return "/api.fcgi", nil },
			expectedError: ErrShortPath,
		},
		{
			name:          "CONFIG_BASE_PATH set but shallow (propagates ErrShortPath)",
			envLookup:     func(string) string { return "/home/pacs/xyz00/api.fcgi" },
			getExecutable: func() (string, error) { return "/tmp/go-build/whatever/exe/main", nil },
			expectedError: ErrShortPath,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := domainByExecutable(tc.envLookup, tc.getExecutable)

			if err == nil {
				t.Error("Expected error but got nil")
			}

			if d != nil {
				t.Error("Expected nil domain but got value")
			}

			if err != tc.expectedError {
				t.Errorf("Expected error %v but got %v", tc.expectedError, err)
			}
		})
	}

	noEnv = func(string) string { return "" }

	for _, tc := range []struct {
		name          string
		envLookup     func(string) string
		getExecutable func() (string, error)
		wantDomain    string
		wantPACErr    error
	}{
		{
			name:          "CONFIG_BASE_PATH to dev fastcgi-ssl binary",
			envLookup:     func(string) string { return "/srv/doms/example.com/fastcgi-ssl/api.fcgi" },
			getExecutable: func() (string, error) { return "/tmp/go-build/main", nil },
			wantDomain:    "example.com",
			wantPACErr:    ErrNoPAC,
		},
		{
			name:          "CONFIG_BASE_PATH to dev doms directory",
			envLookup:     func(string) string { return "/srv/doms/example.com" },
			getExecutable: func() (string, error) { return "/tmp/go-build/main", nil },
			wantDomain:    "example.com",
			wantPACErr:    ErrNoPAC,
		},
		{
			name:          "executable in dev doms tree (no env)",
			envLookup:     noEnv,
			getExecutable: func() (string, error) { return "/srv/doms/example.com/fastcgi-ssl/api.fcgi", nil },
			wantDomain:    "example.com",
			wantPACErr:    ErrNoPAC,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := domainByExecutable(tc.envLookup, tc.getExecutable)
			if err != nil {
				t.Fatalf("Expected no error but got: %v", err)
			}
			if d == nil {
				t.Fatal("Expected domain but got nil")
			}
			if got := d.Domain(); got != tc.wantDomain {
				t.Errorf("Expected domain %q but got %q", tc.wantDomain, got)
			}
			if _, err := d.PAC(); err != tc.wantPACErr {
				t.Errorf("Expected PAC error %v but got %v", tc.wantPACErr, err)
			}
		})
	}
}
