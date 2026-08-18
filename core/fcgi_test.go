package core

import (
	"testing"
)

// TestIsFCGI pins the FastCGI detection contract: the parent directory of
// the executable must have a base name starting with "fastcgi". This rule
// applies to Hostsharing (where Apache aliases /fastcgi-bin/ to
// ~/doms/<host>/fastcgi/) AND to any other deployment where the binary sits
// under a fastcgi-spawning parent directory — so the check itself is
// env-agnostic and belongs in core, not hostsharing.
//
// Moved from hostsharing/path_test.go because the function does not require
// any Hostsharing-specific path layout.
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
		// VM/Caddy style: binary in a fastcgi-spawning parent still triggers true.
		{"/srv/myapp/fastcgi/api", true},
		{"/srv/myapp/fastcgi-ssl/api", true},
		{"/opt/myapp/bin/api", false},
	} {
		if got := isFCGI(func() (string, error) { return tc.path, nil }); got != tc.expected {
			t.Errorf("Expected %v for %v but got %v", tc.expected, tc.path, got)
		}
	}
}
