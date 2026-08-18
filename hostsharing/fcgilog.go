package hostsharing

import (
	"fmt"
	"path/filepath"
	"strings"
)

// FcgiLogFile resolves the FastCGI log file path for the given executable
// path, or "" if no domain can be derived from the path.
//
// On Hostsharing, Apache aliases /fastcgi-bin/ to ~/doms/<host>/fastcgi/,
// so a FastCGI binary's executable path encodes the domain. The log file
// is placed next to the domain's other log files in
// ~/doms/<host>/var/<binary-basename>.log.
//
// Returns "" when the path is too shallow to derive a domain (i.e., not
// running under a Hostsharing-style layout) — callers should fall back to
// stdout in that case.
//
// Used by server.RequestLogger. Test seam lives in the test file via direct
// argument injection (no production wrapper needed — DRY).
func FcgiLogFile(exePath string) (string, error) {
	domain, err := ParseDomain(exePath)
	if err != nil {
		if err == ErrShortPath {
			return "", nil
		}
		return "", err
	}
	b := strings.TrimSuffix(filepath.Base(exePath), ".fcgi")
	return fmt.Sprintf("%s/%s.log", domain.LogDir(), b), nil
}
