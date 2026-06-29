package hostsharing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrShortPath is returned when a path lacks enough components to identify
// a PAC, user, and domain (requires at least 7 path segments for domains,
// 5 for users, or 3 for PAC-only paths).
var ErrShortPath = fmt.Errorf("cannot detect PAC/user/domain from path")

type user struct {
	pac  string
	user *string
}

func (u *user) User() string {
	if u.user != nil {
		return fmt.Sprintf("%s-%s", u.pac, *u.user)
	}
	return u.pac
}

// Home returns the home directory path for the user.
// For PAC users, it returns /home/pacs/{pac}/users/{user}.
// For PAC-only users, it returns /home/pacs/{pac}.
func (u *user) Home() string {
	if u.user != nil {
		return fmt.Sprintf("/home/pacs/%s/users/%s", u.pac, *u.user)
	}
	return fmt.Sprintf("/home/pacs/%s", u.pac)
}

func (u *user) LogDir() string {
	return fmt.Sprintf("%s/var", u.Home())
}

func (u *user) ConfigDir() string {
	return fmt.Sprintf("%s/etc", u.Home())
}

// PAC returns the Web-Paket prefix (e.g. "xyz00"), independent of any
// Domain-Admin or Email-User sub-account name.
func (u *user) PAC() string {
	return u.pac
}

type domain struct {
	user
	domain string
}

// Domain returns the doms hostname (e.g. "example.org") — the directory
// name under .../doms/ where this domain's config, logs, and data live.
func (d *domain) Domain() string {
	return d.domain
}

// DomsDir returns the .../doms/{hostname} directory for this domain,
// without trailing "/etc", "/var", or "/data". It mirrors the layout of
// Home() — pac-only paths drop the /users/{u} segment.
func (d *domain) DomsDir() string {
	return fmt.Sprintf("%s/doms/%s", d.Home(), d.domain)
}

func (d *domain) ConfigDir() string {
	return fmt.Sprintf("%s/doms/%s/etc", d.Home(), d.domain)
}

func (d *domain) LogDir() string {
	return fmt.Sprintf("%s/doms/%s/var", d.Home(), d.domain)
}

func (d *domain) DataDir() string {
	return fmt.Sprintf("%s/doms/%s/data", d.Home(), d.domain)
}

func ParseDomain(p string) (*domain, error) {
	if p == "" {
		return nil, ErrShortPath
	}
	xs := strings.Split(strings.Trim(p, "/"), "/")
	if len(xs) < 7 {
		return nil, ErrShortPath
	}

	u, err := ParseUser(p)
	if err != nil {
		return nil, err
	}
	return &domain{*u, xs[6]}, nil
}

func ParseUser(p string) (*user, error) {
	if p == "" {
		return nil, ErrShortPath
	}
	xs := strings.Split(strings.Trim(p, "/"), "/")
	if len(xs) < 3 {
		return nil, ErrShortPath
	}
	if len(xs) < 5 {
		return &user{pac: xs[2]}, nil
	}
	return &user{pac: xs[2], user: &xs[4]}, nil
}

// domainByWorkingDir resolves the domain from the current working directory.
func domainByWorkingDir(getwd func() (dir string, err error)) (*domain, error) {
	dir, err := getwd()
	if err != nil {
		return nil, err
	}
	return ParseDomain(dir)
}

// DomainByWorkingDir returns the domain parsed from the current working directory.
// Returns ErrShortPath if the path lacks enough components to identify PAC, user, and domain.
//
// Deprecated: Use [DomainByExecutable]. It resists startup `chdir` and is used internally
// by [hostsharing.ReadInConfig] and the database package. Will be removed in v2.
func DomainByWorkingDir() (*domain, error) {
	return domainByWorkingDir(os.Getwd)
}

// domainByExecutable resolves the domain from CONFIG_BASE_PATH first, then the executable's directory.
// Both seams are injected for testability.
//
// CONFIG_BASE_PATH can be a binary path (e.g., `/home/pacs/.../api.fcgi`) or a directory
// (e.g., `/home/pacs/.../doms/example.com`). We parse it as-is first, then try its parent
// directory if ErrShortPath occurs. Other parse errors propagate immediately.
func domainByExecutable(envLookup func(string) string, getExecutable func() (string, error)) (*domain, error) {
	if base := envLookup("CONFIG_BASE_PATH"); base != "" {
		d, err := ParseDomain(base)
		if err == nil {
			return d, nil
		}
		if err != ErrShortPath {
			return nil, err
		}
		// Try the parent directory: env var may point at a binary file.
		d, err = ParseDomain(filepath.Dir(base))
		if err != nil && err != ErrShortPath {
			return nil, err
		}
		if d != nil {
			return d, nil
		}
		// Fall through to the executable branch below.
	}
	exe, err := getExecutable()
	if err != nil {
		return nil, err
	}
	return ParseDomain(filepath.Dir(exe))
}

// DomainByExecutable returns the domain parsed from the current executable's directory path.
//
// Resolution order:
//  1. If CONFIG_BASE_PATH is set, parse it as-is; on ErrShortPath, parse its parent directory.
//     This lets local dev point to a binary (…/api.fcgi) or directory (…/doms/example.com).
//  2. Otherwise, parse the executable's directory.
//
// Returns ErrShortPath if no source has enough path components for PAC/user/domain.
// If CONFIG_BASE_PATH is set but invalid, ErrShortPath propagates to signal the error.
func DomainByExecutable() (*domain, error) {
	return domainByExecutable(os.Getenv, os.Executable)
}

func isFCGI(fn func() (string, error)) bool {
	r, err := fn()
	if err != nil {
		return false
	}
	dir := filepath.Base(filepath.Dir(r))
	return strings.HasPrefix(dir, "fastcgi")
}

// IsFCGI checks if the current executable is running in a FastCGI environment
// by examining the executable path for a "fastcgi" directory component.
func IsFCGI() bool {
	return isFCGI(os.Executable)
}
