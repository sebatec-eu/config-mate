package core

import (
	"os"
	"path/filepath"
)

// Test seams: package-level indirection over os calls.
var (
	getenv         = os.Getenv
	executablePath = os.Executable
)

// XdgConfigHome returns $XDG_CONFIG_HOME, falling back to $HOME/.config,
// or "" when neither is set.
func XdgConfigHome() string {
	if dir := getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}
	if home := getenv("HOME"); home != "" {
		return filepath.Join(home, ".config")
	}
	return ""
}

// XdgConfigDirs returns the viper search dirs in precedence order:
// [<base>/<appName>, <base>], or nil when the base is empty.
func XdgConfigDirs(appName string) []string {
	base := XdgConfigHome()
	if base == "" {
		return nil
	}
	return []string{filepath.Join(base, appName), base}
}

// Unexported alias kept for legacy tests.
func xdgConfigHome() string { return XdgConfigHome() }
