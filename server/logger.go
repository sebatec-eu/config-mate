package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/sebatec-eu/config-mate/v2/core"
	"github.com/sebatec-eu/config-mate/v2/hostsharing"
)

// logWriter returns the io.Writer that RequestLogger should write to.
//
// Precedence:
//  1. When core.IsFCGI() is true, the executable sits under a fastcgi/
//     parent directory (Hostsharing Apache alias). We resolve the per-domain
//     log file via hostsharing.FcgiLogFile. If that fails or the file
//     can't be opened, we fall back to stdout so logging never blocks the
//     request.
//  2. Otherwise stdout — the right default for local dev, CI, and VM
//     deployments where stdout is collected by the orchestrator.
func logWriter() io.Writer {
	if !core.IsFCGI() {
		return os.Stdout
	}
	exePath, err := os.Executable()
	if err != nil {
		return os.Stdout
	}
	logFile, err := hostsharing.FcgiLogFile(exePath)
	if err != nil || logFile == "" {
		return os.Stdout
	}
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return os.Stdout
	}
	return f
}

// RequestLogger returns an HTTP middleware that logs requests using structured logging.
// It sets up a JSON logger configured for Google Cloud Logging schema,
// and logs request/response details including optional requestID from the request context.
// Certain static asset types (css, js, fonts, etc.) are excluded from logging.
func RequestLogger() func(next http.Handler) http.Handler {
	serviceName, err := core.ServiceName()
	if err != nil {
		panic(fmt.Errorf("cannot detect environemnt: %e", err))
	}

	logger := slog.New(slog.NewJSONHandler(logWriter(), &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})).With(
		slog.String("service", serviceName),
		slog.String("version", "latest"),
	)

	slog.SetDefault(logger)

	return httplog.RequestLogger(logger, &httplog.Options{
		Level:         slog.LevelInfo,
		RecoverPanics: true,
		Schema:        httplog.SchemaGCP,
		Skip: func(r *http.Request, respStatus int) bool {
			urlFormat, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
			switch urlFormat {
			case "css", "js", "woff2", "ico", "wasm", "svg", "json":
				return true
			default:
				return false
			}
		},
		LogExtraAttrs: func(r *http.Request, reqBody string, respStatus int) []slog.Attr {
			attrs := []slog.Attr{}
			rID := middleware.GetReqID(r.Context())
			if rID != "" {
				attrs = append(attrs, slog.String("requestID", rID))
			}
			return attrs
		},
	})
}
