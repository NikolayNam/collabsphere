package logger

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

func With(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

func From(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(ctxKey{}).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return l
}
