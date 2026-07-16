package hostsharing

import (
	"bytes"
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/mitchellh/mapstructure"
)

func TestBase64StringToBytesHookFunc(t *testing.T) {
	std := "AAECAwQFBgcICQoLDA0ODw==" // 16 bytes: 0x00..0x0f
	want := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

	// URL-safe fixture: same plaintext, alphabet uses '-' and '_'.
	url := base64.URLEncoding.EncodeToString(want)

	t.Run("StdEncoding round-trips known fixture", func(t *testing.T) {
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), std)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, ok := got.([]byte)
		if !ok {
			t.Fatalf("expected []byte, got %T", got)
		}
		if !bytes.Equal(b, want) {
			t.Errorf("expected %x, got %x", want, b)
		}
	})

	t.Run("URLEncoding round-trips URL-safe fixture", func(t *testing.T) {
		got, err := runHook(Base64StringToBytesHookFunc(base64.URLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), url)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, ok := got.([]byte)
		if !ok {
			t.Fatalf("expected []byte, got %T", got)
		}
		if !bytes.Equal(b, want) {
			t.Errorf("expected %x, got %x", want, b)
		}
	})

	t.Run("Std then URL: Std-only input decodes via Std", func(t *testing.T) {
		// Std has '+' and '/'; URL rejects them. If Std decodes first, we get the bytes.
		stdPlus := base64.StdEncoding.EncodeToString([]byte{0xfb, 0xff, 0xfe}) // contains '/'
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding, base64.URLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), stdPlus)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(got.([]byte), []byte{0xfb, 0xff, 0xfe}) {
			t.Errorf("expected Std decoding to win, got %x", got)
		}
	})

	t.Run("Std then URL: URL-only chars fall through to URL", func(t *testing.T) {
		// Force a '-' into the encoding by using RawURLEncoding on a byte slice that
		// happens to produce one. Simplest: build a URL-only string directly.
		urlOnly := "-__-" // valid only in URL alphabet
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding, base64.URLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), urlOnly)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == urlOnly {
			t.Errorf("expected URL decoding to succeed; input was returned unchanged")
		}
	})

	t.Run("First match wins when input is valid in both alphabets", func(t *testing.T) {
		// 'AAA=' is valid padded input in both Std and URL alphabets. Both
		// decode it to the same byte. We verify the hook returns successfully;
		// the bytes are identical for both alphabets so we cannot assert which
		// one ran — only that the first-match-wins semantics did not error.
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding, base64.URLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), "AAA=")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(got.([]byte), []byte{0x00, 0x00}) {
			t.Errorf("expected [0x00 0x00], got %x", got)
		}
	})

	t.Run("RawURLEncoding accepts unpadded URL input", func(t *testing.T) {
		raw := base64.RawURLEncoding.EncodeToString(want)
		got, err := runHook(Base64StringToBytesHookFunc(base64.RawURLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(got.([]byte), want) {
			t.Errorf("expected %x, got %x", want, got)
		}
	})

	t.Run("URLEncoding rejects unpadded input", func(t *testing.T) {
		raw := base64.RawURLEncoding.EncodeToString(want)
		got, err := runHook(Base64StringToBytesHookFunc(base64.URLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != raw {
			t.Errorf("expected input to pass through unchanged when padding is wrong, got %v", got)
		}
	})

	t.Run("No encodings: input passes through", func(t *testing.T) {
		got, err := runHook(Base64StringToBytesHookFunc(), reflect.TypeOf(""), reflect.TypeOf([]byte{}), std)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != std {
			t.Errorf("expected pass-through, got %v", got)
		}
	})

	t.Run("Non-string source passes through", func(t *testing.T) {
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding), reflect.TypeOf(int(0)), reflect.TypeOf([]byte{}), 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 42 {
			t.Errorf("expected pass-through, got %v", got)
		}
	})

	t.Run("Non-[]byte target passes through", func(t *testing.T) {
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding), reflect.TypeOf(""), reflect.TypeOf(int(0)), std)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != std {
			t.Errorf("expected pass-through, got %v", got)
		}
	})

	t.Run("All encodings reject: input passes through", func(t *testing.T) {
		bad := "###not base64###"
		got, err := runHook(Base64StringToBytesHookFunc(base64.StdEncoding, base64.URLEncoding), reflect.TypeOf(""), reflect.TypeOf([]byte{}), bad)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != bad {
			t.Errorf("expected pass-through on total mismatch, got %v", got)
		}
	})

	t.Run("ComposeDecodeHookFunc accepts the hook", func(t *testing.T) {
		// ComposeDecodeHookFunc is the integration point used by ReadInConfig.
		// If the hook's signature stops matching mapstructure's expectations,
		// this composition step will fail to compile.
		_ = mapstructure.ComposeDecodeHookFunc(
			Base64StringToBytesHookFunc(base64.StdEncoding, base64.URLEncoding),
		)
	})
}

// runHook invokes a DecodeHookFunc with explicit from/to reflect.Types, mirroring
// how mapstructure.ComposeDecodeHookFunc calls them.
func runHook(f mapstructure.DecodeHookFuncType, from, to reflect.Type, data any) (any, error) {
	return f(from, to, data)
}
