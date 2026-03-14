package requestctx

import "context"

type ctxKey struct{}

var key ctxKey

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, key, requestID)
}

func RequestID(ctx context.Context) (string, bool) {
	v := ctx.Value(key)
	if v == nil {
		return "", false
	}
	id, ok := v.(string)
	if !ok || id == "" {
		return "", false
	}
	return id, true
}
