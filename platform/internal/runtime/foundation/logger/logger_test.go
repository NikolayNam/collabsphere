package logger

import (
	"log/slog"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		raw  string
		want slog.Level
	}{
		{raw: "DEBUG", want: slog.LevelDebug},
		{raw: "warn", want: slog.LevelWarn},
		{raw: "ERROR", want: slog.LevelError},
		{raw: "unknown", want: slog.LevelInfo},
		{raw: "", want: slog.LevelInfo},
	}

	for _, tt := range tests {
		if got := ParseLevel(tt.raw); got != tt.want {
			t.Fatalf("ParseLevel(%q) = %v, want %v", tt.raw, got, tt.want)
		}
	}
}
