package logger

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"server/config"
	"strings"
	"time"
)

type CustomLogHandler struct {
	level slog.Level
	attrs map[string]interface{}
	group string
}

func NewCustomLogHandler(level slog.Level) *CustomLogHandler {
	return &CustomLogHandler{
		level: level,
		attrs: make(map[string]interface{}),
	}
}

func (h *CustomLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

//nolint:gocritic // The slog.Record struct triggers hugeParam, but we don't control the interface (it's a standard library one)
func (h *CustomLogHandler) Handle(_ context.Context, record slog.Record) error {
	if !h.Enabled(context.TODO(), record.Level) {
		return nil
	}

	message := fmt.Sprintf("[%s] [%s] %s", time.Now().Format(time.RFC3339), record.Level, record.Message)

	// Append attributes
	for k, v := range h.attrs {
		message += fmt.Sprintf(" | %s=%v", k, v)
	}

	// Log to standard output
	log.Println(message)
	return nil
}

func (h *CustomLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := *h
	newHandler.attrs = make(map[string]interface{})

	for k, v := range h.attrs {
		newHandler.attrs[k] = v
	}

	for _, attr := range attrs {
		newHandler.attrs[attr.Key] = attr.Value.Any()
	}
	return &newHandler
}

func (h *CustomLogHandler) WithGroup(name string) slog.Handler {
	newHandler := *h
	newHandler.group = name
	return &newHandler
}

func ParseLogLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func New() *slog.Logger {
	levelStr := config.AppConfig.Logging.Level
	level := ParseLogLevel(levelStr)
	handler := NewCustomLogHandler(level)
	return slog.New(handler)
}
