package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Logger wraps slog.Logger with project-specific defaults.
type Logger struct {
	inner *slog.Logger
}

// New creates a logger configured for the current environment.
// DEBUG=true enables debug level unless LOG_LEVEL is explicitly set.
func New(debug bool) *Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	if envLevel := strings.TrimSpace(os.Getenv("LOG_LEVEL")); envLevel != "" {
		level = parseLevel(envLevel, level)
	}

	format := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT")))
	addSource := os.Getenv("LOG_SOURCE") == "true"

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	}

	var handler slog.Handler
	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	base := slog.New(handler).With("service", "twitterx-bot")
	slog.SetDefault(base)

	return &Logger{inner: base}
}

// Default returns a wrapper around slog.Default without reconfiguring it.
func Default() *Logger {
	return &Logger{inner: slog.Default()}
}

// With returns a logger with the supplied structured fields.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{inner: l.logger().With(args...)}
}

// Debug logs at DEBUG level.
func (l *Logger) Debug(msg string, args ...any) {
	l.logger().Debug(msg, args...)
}

// Info logs at INFO level.
func (l *Logger) Info(msg string, args ...any) {
	l.logger().Info(msg, args...)
}

// Warn logs at WARN level.
func (l *Logger) Warn(msg string, args ...any) {
	l.logger().Warn(msg, args...)
}

// Error logs at ERROR level.
func (l *Logger) Error(msg string, args ...any) {
	l.logger().Error(msg, args...)
}

func (l *Logger) logger() *slog.Logger {
	if l == nil || l.inner == nil {
		return slog.Default()
	}
	return l.inner
}

func parseLevel(value string, fallback slog.Level) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return fallback
	}
}
