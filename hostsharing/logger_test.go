package hostsharing

import "testing"

func TestFcgiLogFile(t *testing.T) {
	for _, tc := range []struct {
		path    string
		logFile string
	}{
		{"/home/pacs/xyz00/users/example/doms/example.com/fastcgi-ssl/api.fcgi", "/home/pacs/xyz00/users/example/doms/example.com/var/api.log"},
		{"/home/pacs/xyz00/users/example/doms/example.com/fastcgi-ssl/foobar.fcgi", "/home/pacs/xyz00/users/example/doms/example.com/var/foobar.log"},
		{"/home/pacs/xyz00/users/example/doms/example.com/fastcgi/foobar.fcgi", "/home/pacs/xyz00/users/example/doms/example.com/var/foobar.log"},
	} {
		if got, _ := fcgiLogFile(func() (string, error) { return tc.path, nil }); got != tc.logFile {
			t.Errorf("Expected %v for %v but got %v", tc.logFile, tc.path, got)
		}
	}
}

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
			got, err := fcgiLogFile(func() (string, error) { return tc.path, nil })
			if err != nil {
				t.Errorf("Expected nil error but got: %v", err)
			}
			if got != "" {
				t.Errorf("Expected empty log path but got: %q", got)
			}
		})
	}
}
