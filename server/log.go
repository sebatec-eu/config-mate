package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
)

// LogInfo logs an information-level message with the request context.
// If a request ID is present in the context, it is included in the log record.
func LogInfo(ctx context.Context, format string, args ...any) {
	rID := middleware.GetReqID(ctx)
	if rID != "" {
		slog.InfoContext(ctx, fmt.Sprintf(format, args...), "requestID", middleware.GetReqID(ctx))
	} else {
		slog.InfoContext(ctx, fmt.Sprintf(format, args...))
	}
}

// LogWarn logs a warning-level message with the request context.
// If a request ID is present in the context, it is included in the log record.
func LogWarn(ctx context.Context, format string, args ...any) {
	rID := middleware.GetReqID(ctx)
	if rID != "" {
		slog.WarnContext(ctx, fmt.Sprintf(format, args...), "requestID", middleware.GetReqID(ctx))
	} else {
		slog.WarnContext(ctx, fmt.Sprintf(format, args...))
	}
}

// LogError logs an error-level message with the request context.
// If a request ID is present in the context, it is included in the log record.
func LogError(ctx context.Context, format string, args ...any) {
	rID := middleware.GetReqID(ctx)
	if rID != "" {
		slog.ErrorContext(ctx, fmt.Sprintf(format, args...), "requestID", middleware.GetReqID(ctx))
	} else {
		slog.ErrorContext(ctx, fmt.Sprintf(format, args...))
	}
}
