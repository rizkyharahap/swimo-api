package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Init sets slog default logger with file + stderr writers.
// level: "debug" | "info" | "warn" | "error"
// format: "json" | "text"
// If filePath is empty, logs only to stderr.
func Init(appName, level, format, filepath string, addSource bool, extraAttrs ...slog.Attr) (*slog.Logger, func(), error) {
	var lvl slog.Level

	// Parse level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	// Writers: stderr + optional file
	var file *os.File
	writers := []io.Writer{os.Stderr}
	if filepath != "" {
		f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			// fallback: still return stderr-only logger
			logger := slog.New(newHandler(format, io.MultiWriter(writers...), lvl, addSource))
			logger.Error("failed to open log file, fallback to stderr", slog.String("path", filepath))
			slog.SetDefault(logger)
			return logger, func() {}, err
		}
		file = f
		writers = append(writers, file)
	}
	w := io.MultiWriter(writers...)

	handler := newHandler(format, w, lvl, addSource)
	logger := slog.New(handler)

	// default attrs (app name etc.)
	attrs := []any{slog.String("app", appName)}
	for _, a := range extraAttrs {
		attrs = append(attrs, a)
	}
	logger = logger.With(attrs...)
	slog.SetDefault(logger)

	cleanup := func() {
		if file != nil {
			_ = file.Sync()
			_ = file.Close()
		}
	}
	return logger, cleanup, nil
}

func newHandler(format string, w io.Writer, lvl slog.Level, addSource bool) slog.Handler {
	opts := &slog.HandlerOptions{
		Level:     lvl,
		AddSource: addSource,
	}
	switch strings.ToLower(format) {
	case "text":
		return slog.NewTextHandler(w, opts)
	default: // json
		return slog.NewJSONHandler(w, opts)
	}
}
