package dbmodel

import (
	"time"

	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/persistence/gorm/model"
	"github.com/google/uuid"
)

type RefreshSession struct {
	model.UUIDPK
	model.Timestamps

	AccountID uuid.UUID `gorm:"column:account_id;type:uuid;not null;index"`

	TokenHash string `gorm:"column:token_hash;type:text;not null;uniqueIndex"`

	UserAgent *string `gorm:"column:user_agent;type:text"`
	IP        *string `gorm:"column:ip;type:varchar(64)"`

	ExpiresAt time.Time  `gorm:"column:expires_at;type:timestamptz;not null;index"`
	RevokedAt *time.Time `gorm:"column:revoked_at;type:timestamptz"`
}

func (RefreshSession) TableName() string { return "auth.refresh_sessions" }

