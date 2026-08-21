package core

import (
	"path/filepath"
	"testing"
)

func TestXdgConfigDirs(t *testing.T) {
	tests := []struct {
		xdg, home string
		appName   string
		want      []string
	}{
		{"/custom/xdg", "/home/me", "myapp", []string{"/custom/xdg/myapp", "/custom/xdg"}},
		{"", "/home/me", "myapp", []string{"/home/me/.config/myapp", "/home/me/.config"}},
		{"", "", "myapp", nil},
		{"/base", "/home/me", "my-app", []string{filepath.Join("/base", "my-app"), "/base"}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.xdg+"|"+tc.home+"|"+tc.appName, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", tc.xdg)
			t.Setenv("HOME", tc.home)

			got := XdgConfigDirs(tc.appName)
			if len(got) != len(tc.want) {
				t.Fatalf("len: got %d, want %d (%v)", len(got), len(tc.want), got)
			}
			for i, w := range tc.want {
				if got[i] != w {
					t.Errorf("[%d] got %q, want %q", i, got[i], w)
				}
			}
		})
	}
}
