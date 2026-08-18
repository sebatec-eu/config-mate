package hostsharing

import "testing"

// TestFcgiLogFile pins the FastCGI log file path contract for the exported
// FcgiLogFile helper used by server.RequestLogger. The unexported
// fcgiLogFile wrapper was removed (DRY): tests now call the exported form
// directly.
func TestFcgiLogFile(t *testing.T) {
	for _, tc := range []struct {
		path    string
		logFile string
	}{
		{"/home/pacs/xyz00/users/example/doms/example.com/fastcgi-ssl/api.fcgi", "/home/pacs/xyz00/users/example/doms/example.com/var/api.log"},
		{"/home/pacs/xyz00/users/example/doms/example.com/fastcgi-ssl/foobar.fcgi", "/home/pacs/xyz00/users/example/doms/example.com/var/foobar.log"},
		{"/home/pacs/xyz00/users/example/doms/example.com/fastcgi/foobar.fcgi", "/home/pacs/xyz00/users/example/doms/example.com/var/foobar.log"},
	} {
		got, err := FcgiLogFile(tc.path)
		if err != nil {
			t.Errorf("Expected no error for %v but got: %v", tc.path, err)
		}
		if got != tc.logFile {
			t.Errorf("Expected %v for %v but got %v", tc.logFile, tc.path, got)
		}
	}
}

// TestFcgiLogFileShallowPathFallback: when the path is too shallow to derive
// a Hostsharing domain, FcgiLogFile must return "" with a nil error so
// callers can fall back to stdout.
func TestFcgiLogFileShallowPathFallback(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
	}{
		{"empty executable path", ""},
		{"executable at filesystem root", "/api.fcgi"},
		{"executable in pac dir (too shallow)", "/home/pacs/xyz00/api.fcgi"},
		{"executable in users dir (too shallow)", "/home/pacs/xyz00/users/api.fcgi"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FcgiLogFile(tc.path)
			if err != nil {
				t.Errorf("Expected nil error but got: %v", err)
			}
			if got != "" {
				t.Errorf("Expected empty log path but got: %q", got)
			}
		})
	}
}
