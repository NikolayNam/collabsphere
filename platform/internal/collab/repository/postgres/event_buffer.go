package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/google/uuid"
)

type eventBufferRow struct {
	ID        uuid.UUID `gorm:"column:id"`
	ChannelID uuid.UUID `gorm:"column:channel_id"`
	EventType string    `gorm:"column:event_type"`
	Payload   []byte    `gorm:"column:payload"`
	CreatedAt string    `gorm:"column:created_at"`
}

func (r *Repo) StoreRealtimeEvent(ctx context.Context, event collabdomain.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return r.dbFrom(ctx).WithContext(ctx).Exec(
		`INSERT INTO collab.realtime_event_buffer (id, channel_id, event_type, payload) VALUES ($1, $2, $3, $4::jsonb)`,
		uuid.New(), event.ChannelID, event.Type, payload,
	).Error
}

func (r *Repo) DrainRealtimeEvents(ctx context.Context, limit int) ([]collabdomain.Event, []uuid.UUID, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []eventBufferRow
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.realtime_event_buffer").
		Order("created_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, nil, err
	}
	events := make([]collabdomain.Event, 0, len(rows))
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		var event collabdomain.Event
		if err := json.Unmarshal(row.Payload, &event); err != nil {
			continue
		}
		events = append(events, event)
		ids = append(ids, row.ID)
	}
	return events, ids, nil
}

func (r *Repo) DeleteRealtimeEvents(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.dbFrom(ctx).WithContext(ctx).
		Table("collab.realtime_event_buffer").
		Where("id IN ?", ids).
		Delete(nil).Error
}
