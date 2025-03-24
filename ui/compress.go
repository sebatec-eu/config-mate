package ui

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// Compress is a middleware that compresses responds body.
// The only difference of `github.com/go-chi/chi/v5/middleware#Compress` is the default value.
// In order to be able to use this middleware, ensure to set url format with `github.com/go-chi/chi/v5/middleware#URLFormat`
func Compress() func(next http.Handler) http.Handler {
	return middleware.Compress(5, "text/html", "text/css", "text/javascript", "application/javascript",
		"application/x-javascript", "application/json", "image/svg+xml", "image/vnd.microsoft.icon", "font/woff2", "application/wasm")
}
