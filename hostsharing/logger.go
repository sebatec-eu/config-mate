package hostsharing

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/httplog/v2"
)

func fcgiLogFile(fn func() (string, error)) (string, error) {
	exePath, err := fn()
	if err != nil {
		panic(fmt.Errorf("failed to get my own path: %e", err))
	}

	domain, err := ParseDomain(exePath)
	if err != nil {
		panic(fmt.Errorf("failed to get my own path: %e", err))
	}
	b := strings.TrimSuffix(filepath.Base(exePath), ".fcgi")
	return fmt.Sprintf("%s/%s.log", domain.LogDir(), b), nil
}

func FcgiRequestLogger() func(next http.Handler) http.Handler {
	if !IsFCGI() {
		panic(ErrNoFcgiEnvironment)
	}
	appName := fcgiAppName(os.Executable)
	return RequestLogger(appName)
}

func RequestLogger(serviceName string) func(next http.Handler) http.Handler {
	opt := httplog.Options{
		JSON:            true,
		LogLevel:        slog.LevelInfo,
		Concise:         true,
		RequestHeaders:  true,
		ResponseHeaders: true,
		TimeFieldFormat: time.RFC3339,
		Tags: map[string]string{
			"version": "latest",
		},
		QuietDownRoutes: []string{},
		QuietDownPeriod: time.Minute,
	}
	if IsFCGI() {
		logFile, err := fcgiLogFile(os.Executable)
		if err != nil {
			panic(fmt.Errorf("failed to get my own path: %e", err))
		}
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			panic(err)
		}
		opt.Writer = f
	}

	return httplog.RequestLogger(httplog.NewLogger(serviceName, opt))
}
