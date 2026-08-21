package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"github.com/sebatec-eu/config-mate/v2/core"
	"github.com/sebatec-eu/config-mate/v2/hostsharing"
	"github.com/spf13/viper"
)

// ReadInConfig loads config into rawVal. App name: [core.ServiceName].
// Missing file is not an error: rawVal keeps its zero values plus any prior
// viper.SetDefault calls (viper ignores mapstructure `default:` tags).
//
// Search order:
//  1. <domain.ConfigDir>/<app> — PAC layout (CONFIG_BASE_PATH honored for dev).
//  2. $XDG_CONFIG_HOME/<app>/<app>.{ext}, then $XDG_CONFIG_HOME/<app>.{ext}
//     (or $HOME/.config fallback).
//  3. $HOME/.<app> (legacy).
//
// File basename must equal <appName>. Extensions: yaml, yml, json, toml,
// properties, props, prop, hcl, tfvars, dotenv, env, ini; extensionless
// <appName> also works because SetConfigType("yaml") is set.
//
// rawVal must be a pointer. fs adds mapstructure decode hooks; when empty,
// defaults are Base64StringToBytesHookFunc(Std, URL), StringToTimeDurationHookFunc,
// StringToSliceHookFunc(",").
func ReadInConfig(rawVal any, fs ...mapstructure.DecodeHookFunc) error {
	appName, err := core.ServiceName()
	if err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(appName)

	if cfgDir, err := hostsharingConfigDir(); err != nil {
		log.Printf("config-mate: PAC detection failed (%v); continuing with XDG fallback", err)
	} else if cfgDir != "" {
		v.AddConfigPath(cfgDir)
	}
	for _, p := range core.XdgConfigDirs(appName) {
		v.AddConfigPath(p)
	}
	if home := os.Getenv("HOME"); home != "" {
		v.AddConfigPath(filepath.Join(home, "."+appName))
	}

	if err := v.ReadInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
		return fmt.Errorf("cannot read config: %w", err)
	}

	if len(fs) <= 0 {
		fs = append(fs,
			core.Base64StringToBytesHookFunc(base64.StdEncoding, base64.URLEncoding),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)
	}

	if err := v.Unmarshal(&rawVal, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(fs...))); err != nil {
		return fmt.Errorf("cannot unmarshal config: %v", err)
	}

	return nil
}

// hostsharingConfigDir returns the PAC layout ConfigDir, or ("", nil) for
// no match. ErrShortPath is treated as a non-match (expected for non-PAC
// deployments). Test seam: DomainByExecutable's *domain is unexported, so
// the seam exposes the ConfigDir string directly.
var hostsharingConfigDir = func() (string, error) {
	d, err := hostsharing.DomainByExecutable()
	if err == hostsharing.ErrShortPath {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if d == nil {
		return "", nil
	}
	return d.ConfigDir(), nil
}
