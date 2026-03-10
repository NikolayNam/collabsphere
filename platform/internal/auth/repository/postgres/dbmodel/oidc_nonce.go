package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type OIDCNonce struct {
	ID           uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	Provider     string     `gorm:"column:provider;type:varchar(64);not null"`
	OAuthStateID uuid.UUID  `gorm:"column:oauth_state_id;type:uuid;not null;uniqueIndex"`
	NonceHash    string     `gorm:"column:nonce_hash;type:text;not null;uniqueIndex"`
	ExpiresAt    time.Time  `gorm:"column:expires_at;type:timestamptz;not null;index"`
	UsedAt       *time.Time `gorm:"column:used_at;type:timestamptz"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;type:timestamptz"`
}

func (OIDCNonce) TableName() string { return "auth.oidc_nonces" }
