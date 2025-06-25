package hostsharing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrShortPath     = fmt.Errorf("cannot dedect anything")
	ErrUnkownAppName = fmt.Errorf("app name unkown")
)

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

type domain struct {
	user
	domain string
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

func DomainByWorkingDir() (*domain, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return ParseDomain(dir)
}

func isFCGI(fn func() (string, error)) bool {
	r, err := fn()
	if err != nil {
		return false
	}
	dir := filepath.Base(filepath.Dir(r))
	return strings.HasPrefix(dir, "fastcgi")
}

func appName(fn func() (string, error)) (string, error) {
	r, err := fn()
	if err != nil {
		return "", ErrUnkownAppName
	}

	return strings.TrimSuffix(filepath.Base(r), ".fcgi"), nil
}

func IsFCGI() bool {
	return isFCGI(os.Executable)
}
