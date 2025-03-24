package ui

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"text/template"
)

func Template(tmplRoot fs.FS, filename string) http.HandlerFunc {
	t, err := template.ParseFS(tmplRoot, "*.tmpl")
	if err != nil {
		log.Fatalln("Cannot read template files: ", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		t.ExecuteTemplate(w, filename, fromContext(r.Context()))
	}
}

type contextTemplateType struct{ string }
type templateData map[string]any

var ctxTempl = contextTemplateType{"tmpl"}

func fromContext(ctx context.Context) *templateData {
	raw, ok := ctx.Value(ctxTempl).(*templateData)
	if !ok {
		return &templateData{}
	}
	return raw
}

func AddTemplateValue(ctx context.Context, key string, val any) {
	data := *fromContext(ctx)
	data[key] = val
}

func InitTemplateContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxTempl, &templateData{"request": r})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
