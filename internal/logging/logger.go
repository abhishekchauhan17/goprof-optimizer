package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Logger is the logging interface used across the service. It intentionally
// mirrors a subset of slog.Logger to keep it simple and easy to adapt.
type Logger interface {
	Debug(msg string, kv ...any)
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Error(msg string, kv ...any)
	With(kv ...any) Logger
}

type stdLogger struct {
	slog *slog.Logger
}

// NewLogger constructs a new structured logger writing to stderr. Level is
// interpreted case-insensitively: "debug", "info", "warn", "error".
func NewLogger(level string) Logger {
	var slogLevel slog.Level

	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		// default to info
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slogLevel,
	})

	return &stdLogger{slog: slog.New(handler)}
}

func (l *stdLogger) Debug(msg string, kv ...any) {
	l.slog.Debug(msg, kv...)
}

func (l *stdLogger) Info(msg string, kv ...any) {
	l.slog.Info(msg, kv...)
}

func (l *stdLogger) Warn(msg string, kv ...any) {
	l.slog.Warn(msg, kv...)
}

func (l *stdLogger) Error(msg string, kv ...any) {
	l.slog.Error(msg, kv...)
}

func (l *stdLogger) With(kv ...any) Logger {
	return &stdLogger{slog: l.slog.With(kv...)}
}

// Noop returns a Logger implementation that discards all logs.
// Useful in tests.
func Noop() Logger {
	return &noopLogger{}
}

type noopLogger struct{}

func (n *noopLogger) Debug(string, ...any) {}
func (n *noopLogger) Info(string, ...any)  {}
func (n *noopLogger) Warn(string, ...any)  {}
func (n *noopLogger) Error(string, ...any) {}
func (n *noopLogger) With(...any) Logger   { return n }
