package postgres

import (
	"context"

	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/google/uuid"
)

// ChannelLoaderAdapter adapts Repo to realtime.ChannelLoader.
type ChannelLoaderAdapter struct {
	repo *Repo
}

// NewChannelLoaderAdapter returns an adapter for realtime channel/message loading.
func NewChannelLoaderAdapter(repo *Repo) *ChannelLoaderAdapter {
	return &ChannelLoaderAdapter{repo: repo}
}

// ListAllChannelIDs implements realtime.ChannelLoader.
func (a *ChannelLoaderAdapter) ListAllChannelIDs(ctx context.Context) ([]uuid.UUID, error) {
	return a.repo.ListAllChannelIDs(ctx)
}

// ListRecentMessagesForChannel implements realtime.ChannelLoader.
func (a *ChannelLoaderAdapter) ListRecentMessagesForChannel(ctx context.Context, channelID uuid.UUID, limit int) ([]collabdomain.Message, error) {
	return a.repo.ListRecentMessagesForChannel(ctx, channelID, limit)
}
