package realtime

import (
	"context"

	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/google/uuid"
)

// ChannelLoader loads channels and their messages from Postgres for Redis broker init.
// Used to preload collab.channels and collab.messages so subscribers get history on connect.
type ChannelLoader interface {
	ListAllChannelIDs(ctx context.Context) ([]uuid.UUID, error)
	ListRecentMessagesForChannel(ctx context.Context, channelID uuid.UUID, limit int) ([]collabdomain.Message, error)
}
