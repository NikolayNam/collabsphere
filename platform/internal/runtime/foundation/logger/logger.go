package logger

import (
	"io"
	"log/slog"
	"os"
)

type Config struct {
	Level     slog.Leveler
	AddSource bool
	Format    string // "json" | "text"
	Output    io.Writer
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

	return slog.New(h)
}
