package server

import (
	"context"
	"log/slog"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
)

// TestLogInfoWarnError pins the request-context plumbing: when the context
// carries a request ID, LogInfo/Warn/Error attach it as "requestID"; when it
// doesn't, they emit the message as-is.
//
// These are pure slog wrappers around middleware.GetReqID; they moved to
// server/ because the rest of the logging surface (RequestLogger) lives there.
func TestLogInfoWarnError(t *testing.T) {
	t.Run("plain context: no requestID attr", func(t *testing.T) {
		var seen slog.Level
		rec := &recorder{}
		slog.SetDefault(slog.New(slog.NewJSONHandler(rec, &slog.HandlerOptions{Level: slog.LevelInfo})))

		LogInfo(context.Background(), "hello %s", "world")
		LogWarn(context.Background(), "warn %d", 42)
		LogError(context.Background(), "err %s", "boom")

		if len(rec.lines) != 3 {
			t.Fatalf("expected 3 log lines, got %d", len(rec.lines))
		}
		for i, l := range rec.lines {
			if contains(l, `"requestID"`) {
				t.Errorf("line %d: did not expect requestID attr, got %s", i, l)
			}
		}
		_ = seen // silence unused
	})

	t.Run("context with request ID: includes requestID attr", func(t *testing.T) {
		rec := &recorder{}
		slog.SetDefault(slog.New(slog.NewJSONHandler(rec, &slog.HandlerOptions{Level: slog.LevelInfo})))

		ctx := context.WithValue(context.Background(), middleware.RequestIDKey, "req-123")

		LogInfo(ctx, "info %s", "ok")
		LogWarn(ctx, "warn %s", "ok")
		LogError(ctx, "err %s", "ok")

		if len(rec.lines) != 3 {
			t.Fatalf("expected 3 log lines, got %d", len(rec.lines))
		}
		for i, l := range rec.lines {
			if !contains(l, `"requestID":"req-123"`) {
				t.Errorf("line %d: expected requestID=req-123, got %s", i, l)
			}
		}
	})
}

type recorder struct {
	lines []string
}

func (r *recorder) Write(b []byte) (int, error) {
	r.lines = append(r.lines, string(b))
	return len(b), nil
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
