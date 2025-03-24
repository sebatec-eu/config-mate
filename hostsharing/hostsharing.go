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

func ReadInConfig(rawVal any, app_name string) error {
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

	if err := viper.Unmarshal(&rawVal, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		base64StringToBytesHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	))); err != nil {
		return fmt.Errorf("cannot unmarshal config: %v", err)
	}

	return nil
}
