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
