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

		if got := u.User(); got != tc.expected {
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

		if got := u.User(); got != tc.expectedUser {
			t.Errorf("Expected %s but got %s", tc.expectedUser, got)
		}

		if got := u.domain; got != tc.expectedDomain {
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
		{user{"xyz00", nil}, "xyz00"},
		{user{"xyz00", &[]string{"example"}[0]}, "xyz00-example"},
		{user{"xyz00", &[]string{"www.example.com"}[0]}, "xyz00-www.example.com"},
	} {
		if got := tc.User(); got != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, got)
		}
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
		{domain{user{"xyz00", nil}, "example.com"}, "/home/pacs/xyz00"},
		{domain{user{"xyz00", &[]string{"example"}[0]}, "example.com"}, "/home/pacs/xyz00/users/example"},
		{domain{user{"xyz00", &[]string{"www.example.com"}[0]}, "example.com"}, "/home/pacs/xyz00/users/www.example.com"},
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
		{domain{user{"xyz00", nil}, "example.com"}, "/home/pacs/xyz00/doms/example.com/etc"},
		{domain{user{"xyz00", &[]string{"example"}[0]}, "example.com"}, "/home/pacs/xyz00/users/example/doms/example.com/etc"},
		{domain{user{"xyz00", &[]string{"www.example.com"}[0]}, "example.com"}, "/home/pacs/xyz00/users/www.example.com/doms/example.com/etc"},
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
		{domain{user{"xyz00", nil}, "example.com"}, "/home/pacs/xyz00/doms/example.com/var"},
		{domain{user{"xyz00", &[]string{"example"}[0]}, "example.com"}, "/home/pacs/xyz00/users/example/doms/example.com/var"},
		{domain{user{"xyz00", &[]string{"www.example.com"}[0]}, "example.com"}, "/home/pacs/xyz00/users/www.example.com/doms/example.com/var"},
	} {
		if got := tc.LogDir(); got != tc.expected {
			t.Errorf("Expected %s but got %s", tc.expected, got)
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

func TestDomainByWorkingDir(t *testing.T) {
	// Test successful cases
	for _, tc := range []struct {
		name           string
		getwd          func() (string, error)
		expectedUser   string
		expectedDomain string
	}{
		{
			name:           "valid path with user and domain",
			getwd:          func() (string, error) { return "/home/pacs/xyz00/users/foobar/doms/example.com", nil },
			expectedUser:   "xyz00-foobar",
			expectedDomain: "example.com",
		},
		{
			name:           "valid path with subdirectories",
			getwd:          func() (string, error) { return "/home/pacs/xyz00/users/foobar/doms/example.com/fastcgi-ssl/api.fcgi", nil },
			expectedUser:   "xyz00-foobar",
			expectedDomain: "example.com",
		},
		{
			name:           "valid path with trailing slash",
			getwd:          func() (string, error) { return "/home/pacs/xyz00/users/foobar/doms/example.com/", nil },
			expectedUser:   "xyz00-foobar",
			expectedDomain: "example.com",
		},
		{
			name:           "valid path with minimum segments",
			getwd:          func() (string, error) { return "/home/pacs/abc/users/def/doms/test.org", nil },
			expectedUser:   "abc-def",
			expectedDomain: "test.org",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := domainByWorkingDir(tc.getwd)
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if d == nil {
				t.Fatal("Expected domain but got nil")
			}

			if got := d.User(); got != tc.expectedUser {
				t.Errorf("Expected user %s but got %s", tc.expectedUser, got)
			}

			if got := d.domain; got != tc.expectedDomain {
				t.Errorf("Expected domain %s but got %s", tc.expectedDomain, got)
			}
		})
	}

	// Test error cases
	for _, tc := range []struct {
		name          string
		getwd         func() (string, error)
		expectedError error
	}{
		{
			name:          "getwd returns error",
			getwd:         func() (string, error) { return "", ErrShortPath },
			expectedError: ErrShortPath,
		},
		{
			name:          "getwd returns empty string",
			getwd:         func() (string, error) { return "", nil },
			expectedError: ErrShortPath,
		},
		{
			name:          "getwd returns path with less than 7 segments",
			getwd:         func() (string, error) { return "/home/pacs/xyz00", nil },
			expectedError: ErrShortPath,
		},
		{
			name:          "getwd returns path with only pac, no user or domain",
			getwd:         func() (string, error) { return "/home/pacs/xyz00/users", nil },
			expectedError: ErrShortPath,
		},
		{
			name:          "getwd returns short path",
			getwd:         func() (string, error) { return "/home/pacs", nil },
			expectedError: ErrShortPath,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := domainByWorkingDir(tc.getwd)

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
}
