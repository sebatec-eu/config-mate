package core

import (
	"encoding/base64"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// Base64StringToBytesHookFunc returns a mapstructure decode hook that turns
// string fields into []byte by trying each encoding in order. The first
// encoding that decodes the string wins; if none decode it, the string is
// returned unchanged. Pass one encoding to accept a single alphabet, several
// to accept a fallback chain (for example StdEncoding then URLEncoding).
// Padding rules are set on each encoding — use RawStdEncoding, RawURLEncoding,
// or Encoding.WithPadding for unpadded or custom-padded input.
func Base64StringToBytesHookFunc(encs ...*base64.Encoding) mapstructure.DecodeHookFuncType {
	return func(f, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf([]byte{}) {
			return data, nil
		}
		s, _ := data.(string)
		for _, enc := range encs {
			if result, err := enc.DecodeString(s); err == nil {
				return result, nil
			}
		}
		return data, nil
	}
}
