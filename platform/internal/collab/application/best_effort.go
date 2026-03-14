package application

import (
	"context"
	"log/slog"
)

func (s *Service) bestEffort(ctx context.Context, operation string, fn func() error) {
	if fn == nil {
		return
	}
	if err := fn(); err != nil {
		slog.WarnContext(ctx, "best-effort operation failed", "operation", operation, "error", err.Error())
	}
}
