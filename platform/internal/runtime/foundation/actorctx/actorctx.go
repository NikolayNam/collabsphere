package actorctx

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey struct{}

var key ctxKey

func WithActorID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, key, id)
}

func ActorID(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(key)
	if v == nil {
		return uuid.UUID{}, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func MustActorID(ctx context.Context) uuid.UUID {
	id, ok := ActorID(ctx)
	if !ok {
		panic("actor id not found in context")
	}
	return id
}
