package hostsharing

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/httplog/v3"
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

func logWriter() io.Writer {
	if IsFCGI() {
		logFile, err := fcgiLogFile(os.Executable)
		if err != nil {
			panic(fmt.Errorf("failed to get my own path: %e", err))
		}
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			panic(err)
		}
		return f
	}
	return os.Stdout
}

func RequestLogger() func(next http.Handler) http.Handler {
	serviceName, err := appName(os.Executable)
	if err != nil {
		panic(fmt.Errorf("cannot detect environemnt: %e", err))
	}

	logger := slog.New(slog.NewJSONHandler(logWriter(), &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})).With(
		slog.String("service", serviceName),
		slog.String("version", "latest"),
	)

	return httplog.RequestLogger(logger, &httplog.Options{
		Level:         slog.LevelInfo,
		RecoverPanics: true,
	})
}
