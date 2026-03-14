package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Level     slog.Leveler
	AddSource bool
	Format    string // "json" | "text"
	Output    io.Writer
	Fields    []any
}

func New(cfg Config) *slog.Logger {
	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	var h slog.Handler
	switch cfg.Format {
	case "text":
		h = slog.NewTextHandler(out, opts)
	default:
		h = slog.NewJSONHandler(out, opts)
	}

	log := slog.New(h)
	if len(cfg.Fields) > 0 {
		log = log.With(cfg.Fields...)
	}
	return log
}

func ParseLevel(raw string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
