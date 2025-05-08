package ui

import (
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func lastModifiedMiddleware(fs http.FileSystem) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f, err := fs.Open(r.URL.Path)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			defer f.Close()

			fi, err := f.Stat()
			if err != nil || fi.IsDir() {
				next.ServeHTTP(w, r)
				return
			}

			modTime := fi.ModTime()
			if t, err := http.ParseTime(r.Header.Get("If-Modified-Since")); err == nil && modTime.Before(t.Add(1*time.Second)) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("Last-Modified", modTime.UTC().Format(http.TimeFormat))
			next.ServeHTTP(w, r)
		})
	}
}

func ServeStaticOrTemplate(staticRoot http.FileSystem, tmplRoot fs.FS) http.Handler {
	stHandler := lastModifiedMiddleware(staticRoot)(http.FileServer(staticRoot))
	tmplHandlerFunc := Template(tmplRoot, "index.html.tmpl")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlFormat, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
		switch urlFormat {
		case "css", "js", "woff2", "ico", "wasm", "svg", "json":
			stHandler.ServeHTTP(w, r)
		case "":
			tmplHandlerFunc(w, r)
		default:
			panic(fmt.Errorf("undefined format case: %v", urlFormat))
		}

	})
}
