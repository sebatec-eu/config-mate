package core

import (
	"os"
	"path/filepath"
)

// getenv is a package-local indirection over os.Getenv so tests can stub it
// without touching the real environment.
var getenv = os.Getenv

// executablePath is a package-local indirection over os.Executable so tests
// can stub it. It is a var (not a const) so test code can swap it.
var executablePath = os.Executable

// XdgConfigHome returns the XDG config home directory: $XDG_CONFIG_HOME if
// set, otherwise $HOME/.config. Returns "" when neither variable is set.
//
// The helper is env-only and has no Hostsharing-specific assumption, so it
// lives in core rather than hostsharing.
func XdgConfigHome() string {
	if dir := getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}
	if home := getenv("HOME"); home != "" {
		return filepath.Join(home, ".config")
	}
	return ""
}

// xdgConfigHome is kept as an unexported alias so existing tests in core can
// continue to reference the private symbol without churn.
func xdgConfigHome() string { return XdgConfigHome() }
