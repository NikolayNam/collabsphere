package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type RefreshSessionToken struct {
	ID        uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	SessionID uuid.UUID  `gorm:"column:session_id;type:uuid;not null;index"`
	TokenHash string     `gorm:"column:token_hash;type:text;not null;uniqueIndex"`
	UsedAt    *time.Time `gorm:"column:used_at;type:timestamptz"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
}

func (RefreshSessionToken) TableName() string { return "auth.refresh_session_tokens" }
