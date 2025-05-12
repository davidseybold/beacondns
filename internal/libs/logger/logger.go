package logger

import (
	"io"
	"log/slog"
	"strings"
)

func NewJSONLogger(level slog.Level, w io.Writer) *slog.Logger {
	jsonHandler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})

	return slog.New(jsonHandler)
}

func NewDiscardLogger() *slog.Logger {
	return slog.New(discardHandler{})
}

func GetLogLevelFromString(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug.Level()
	case "info":
		return slog.LevelInfo.Level()
	case "warn":
		return slog.LevelWarn.Level()
	case "error":
		return slog.LevelError.Level()
	default:
		return slog.LevelInfo.Level()
	}
}
