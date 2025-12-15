// Package logger provides structured logging with level filtering.
package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Logger wraps slog.Logger with application-specific configuration.
type Logger struct {
	logger *slog.Logger
}

// New creates a new logger with the specified level.
// Level must be one of: debug, info, warn, error (case-insensitive).
// Returns error if level is invalid.
func New(level string) (*Logger, error) {
	slogLevel, err := parseLevel(level)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &Logger{logger: logger}, nil
}

// parseLevel converts string level to slog.Level.
func parseLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level %q (must be debug, info, warn, or error)", level)
	}
}

// Debug logs a debug-level message with optional key-value pairs.
func (l *Logger) Debug(msg string, keysAndValues ...any) {
	l.logger.Debug(msg, keysAndValues...)
}

// Info logs an info-level message with optional key-value pairs.
func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.logger.Info(msg, keysAndValues...)
}

// Warn logs a warning-level message with optional key-value pairs.
func (l *Logger) Warn(msg string, keysAndValues ...any) {
	l.logger.Warn(msg, keysAndValues...)
}

// Error logs an error-level message with optional key-value pairs.
func (l *Logger) Error(msg string, keysAndValues ...any) {
	l.logger.Error(msg, keysAndValues...)
}

// With returns a new Logger with additional context fields.
// The returned logger inherits the parent's level and configuration.
func (l *Logger) With(keysAndValues ...any) *Logger {
	return &Logger{
		logger: l.logger.With(keysAndValues...),
	}
}
