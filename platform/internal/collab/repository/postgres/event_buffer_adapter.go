package postgres

import (
	"context"

	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/google/uuid"
)

// EventBufferStoreAdapter adapts Repo to EventBufferStore for Redis fallback.
type EventBufferStoreAdapter struct {
	repo *Repo
}

// NewEventBufferStoreAdapter returns an adapter for realtime event buffer storage.
func NewEventBufferStoreAdapter(repo *Repo) *EventBufferStoreAdapter {
	return &EventBufferStoreAdapter{repo: repo}
}

// Store saves an event to the buffer.
func (a *EventBufferStoreAdapter) Store(ctx context.Context, event collabdomain.Event) error {
	return a.repo.StoreRealtimeEvent(ctx, event)
}

// Drain fetches events from the buffer for publishing to Redis.
func (a *EventBufferStoreAdapter) Drain(ctx context.Context, limit int) ([]collabdomain.Event, []uuid.UUID, error) {
	return a.repo.DrainRealtimeEvents(ctx, limit)
}

// Delete removes published events from the buffer.
func (a *EventBufferStoreAdapter) Delete(ctx context.Context, ids []uuid.UUID) error {
	return a.repo.DeleteRealtimeEvents(ctx, ids)
}
