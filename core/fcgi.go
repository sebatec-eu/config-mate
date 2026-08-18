package core

import (
	"path/filepath"
	"strings"
)

// isFCGI returns true when the executable's parent directory has a base name
// starting with "fastcgi". The injection seam (fn) lets tests feed in
// arbitrary paths without touching os.Executable.
//
// This rule is environment-agnostic: it works on Hostsharing (where Apache
// aliases /fastcgi-bin/ to ~/doms/<host>/fastcgi/) and on any other setup
// where the binary sits under a fastcgi-spawning parent directory. Therefore
// the function lives in core, not in hostsharing.
func isFCGI(fn func() (string, error)) bool {
	r, err := fn()
	if err != nil {
		return false
	}
	dir := filepath.Base(filepath.Dir(r))
	return strings.HasPrefix(dir, "fastcgi")
}

// IsFCGI checks if the current executable is running in a FastCGI
// environment by examining the executable's parent directory name. Returns
// false if the executable path cannot be resolved.
//
// Detection is purely by parent-directory name (the "fastcgi" prefix), which
// is the convention used by Hostsharing's Apache alias /fastcgi-bin/ and by
// any reverse proxy that spawns the binary from a fastcgi-named directory.
func IsFCGI() bool {
	return isFCGI(executablePath)
}
