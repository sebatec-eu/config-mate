package ui

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func ServeStaticOrTemplate(staticRoot http.FileSystem, tmplRoot fs.FS) http.Handler {
	stHandler := http.FileServer(staticRoot)
	tmplHandlerFunc := Template(tmplRoot, "index.html.tmpl")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlFormat, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
		switch urlFormat {
		case "css":
			stHandler.ServeHTTP(w, r)
		case "js":
			stHandler.ServeHTTP(w, r)
		case "woff2":
			stHandler.ServeHTTP(w, r)
		case "ico":
			stHandler.ServeHTTP(w, r)
		case "wasm":
			stHandler.ServeHTTP(w, r)
		case "":
			tmplHandlerFunc(w, r)
		default:
			panic(fmt.Errorf("undefined format case: %v", urlFormat))
		}

	})
}
