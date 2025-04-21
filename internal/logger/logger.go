package logger

import (
	"io"
	"log/slog"
)

func InitLogger(level string, format string, out io.Writer) *slog.Logger {
	var handler slog.Handler

	var sLevel slog.Level

	switch level {
	case "ERROR":
		sLevel = slog.LevelError
	case "INFO":
		sLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     sLevel,
		AddSource: false,
	}

	switch format {
	case "json":
		handler = slog.NewJSONHandler(out, opts)

	case "text":
		handler = slog.NewTextHandler(out, opts)

	default:
		handler = slog.NewJSONHandler(out, opts)
	}

	logger := slog.New(handler)

	return logger
}
