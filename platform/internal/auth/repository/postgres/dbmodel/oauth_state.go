package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type OAuthState struct {
	ID           uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	Provider     string     `gorm:"column:provider;type:varchar(64);not null"`
	StateHash    string     `gorm:"column:state_hash;type:text;not null;uniqueIndex"`
	CodeVerifier string     `gorm:"column:code_verifier;type:text"`
	ReturnTo     string     `gorm:"column:return_to;type:text;not null"`
	Intent       string     `gorm:"column:intent;type:varchar(32);not null"`
	ExpiresAt    time.Time  `gorm:"column:expires_at;type:timestamptz;not null;index"`
	UsedAt       *time.Time `gorm:"column:used_at;type:timestamptz"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;type:timestamptz"`
}

func (OAuthState) TableName() string { return "auth.oauth_states" }
