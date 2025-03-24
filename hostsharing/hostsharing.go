package hostsharing

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/http/fcgi"
	"os"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const defaultHttpPort = "9000"

var ErrNoFcgiEnvironment = fmt.Errorf("no fcgi environment dedected")

func ListenAndServe(handler http.Handler) error {
	if IsFCGI() {
		if err := fcgi.Serve(nil, handler); err != nil {
			return fmt.Errorf("cannot run server: %v", err)
		}
	} else {
		log.Println("Server listening on port ", defaultHttpPort)
		if err := http.ListenAndServe(":"+defaultHttpPort, handler); err != nil {
			return fmt.Errorf("cannot run server: %v", err)
		}
	}
	return nil
}

func base64StringToBytesHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf([]byte{}) {
			return data, nil
		}

		if result, err := base64.StdEncoding.DecodeString(data.(string)); err == nil {
			return result, nil
		}

		return data, nil
	}
}

// It tries to dedect an FCGI environement on the Hostsharing plattform. Usually
// a binary is located under /home/pacs/xyz00/users/example/doms/example.com/fastcgi-ssl/hello.fcgi
// In this case, the app name is "hello". It is used to search for the config file.
func FcgiReadInConfig(rawVal any, fs ...mapstructure.DecodeHookFunc) error {
	if !IsFCGI() {
		return ErrNoFcgiEnvironment
	}
	appName, err := appName(os.Executable)
	if err != nil {
		panic(fmt.Errorf("cannot detect environemnt: %e", err))
	}
	return ReadInConfig(rawVal, appName, fs)
}

func ReadInConfig(rawVal any, app_name string, fs ...mapstructure.DecodeHookFunc) error {
	viper.SetConfigType("yaml")
	cfg, err := os.ReadFile(fmt.Sprintf(".%s.conf", app_name))
	if err != nil {
		domain, err := DomainByWorkingDir()
		if err != nil && err != ErrShortPath {
			panic(err)
		}
		if domain != nil {
			viper.AddConfigPath(fmt.Sprintf("%s/%s", domain.ConfigDir(), app_name))
		}
		viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", app_name))

		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("fatal error config file: %w", err)
		}
	} else {
		viper.ReadConfig(bytes.NewBuffer(cfg))
	}

	if len(fs) <= 0 {
		fs = append(fs,
			base64StringToBytesHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)
	}

	if err := viper.Unmarshal(&rawVal, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(fs))); err != nil {
		return fmt.Errorf("cannot unmarshal config: %v", err)
	}

	return nil
}
