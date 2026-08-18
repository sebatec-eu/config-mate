package server

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"github.com/sebatec-eu/config-mate/v2/core"
	"github.com/sebatec-eu/config-mate/v2/hostsharing"
	"github.com/spf13/viper"
)

// ReadInConfig reads and unmarshals configuration from a file into rawVal.
// The application name comes from [core.ServiceName] (SERVICE_NAME env, else
// executable basename with optional .fcgi stripped).
//
// A missing config file is not an error: ReadInConfig returns nil and rawVal
// retains its zero values plus any viper.SetDefault calls made before this
// function (mapstructure `default:"..."` struct tags are NOT applied by viper).
//
// Search order:
//  1. .<app>.conf in the current working directory (loaded via ReadConfig).
//  2. <domain.ConfigDir>/<app> — e.g. /home/pacs/.../doms/example.com/etc/api
//     (resolved via [hostsharing.DomainByExecutable]; CONFIG_BASE_PATH is
//     honored for local dev).
//  3. $XDG_CONFIG_HOME/<app>, falling back to $HOME/.config/<app>.
//  4. $HOME/.<app> (legacy).
//
// rawVal must be a pointer. fs optionally adds mapstructure decode hooks; when
// empty, defaults are: core.Base64StringToBytesHookFunc(Std, URL),
// mapstructure.StringToTimeDurationHookFunc, mapstructure.StringToSliceHookFunc(",").
func ReadInConfig(rawVal any, fs ...mapstructure.DecodeHookFunc) error {
	appName, err := core.ServiceName()
	if err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(appName)

	if cfg, err := os.ReadFile("." + appName + ".conf"); err == nil {
		if err := v.ReadConfig(bytes.NewBuffer(cfg)); err != nil {
			return fmt.Errorf("cannot read local config: %w", err)
		}
	} else {
		if domain, err := hostsharing.DomainByExecutable(); err != nil && err != hostsharing.ErrShortPath {
			panic(err)
		} else if domain != nil {
			v.AddConfigPath(domain.ConfigDir())
		}
		if xdg := core.XdgConfigHome(); xdg != "" {
			v.AddConfigPath(xdg)
		}
		if home := os.Getenv("HOME"); home != "" {
			v.AddConfigPath(filepath.Join(home, "."+appName))
		}

		if err := v.ReadInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return fmt.Errorf("cannot read config: %w", err)
		}
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
