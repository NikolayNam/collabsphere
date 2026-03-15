package realtime

import (
	"context"

	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/google/uuid"
)

// EventBufferStore persists realtime events when Redis is unavailable.
// RedisBroker uses it as fallback on publish failure and drains to Redis when it recovers.
type EventBufferStore interface {
	Store(ctx context.Context, event collabdomain.Event) error
	Drain(ctx context.Context, limit int) (events []collabdomain.Event, ids []uuid.UUID, err error)
	Delete(ctx context.Context, ids []uuid.UUID) error
}
