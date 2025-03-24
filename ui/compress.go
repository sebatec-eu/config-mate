package ui

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func Compress() func(next http.Handler) http.Handler {
	return middleware.Compress(5, "text/html", "text/css", "text/javascript", "application/javascript",
		"application/x-javascript", "application/json", "image/svg+xml", "image/vnd.microsoft.icon", "font/woff2", "application/wasm")
}
