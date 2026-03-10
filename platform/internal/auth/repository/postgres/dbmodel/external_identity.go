package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type ExternalIdentity struct {
	ID              uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	Provider        string     `gorm:"column:provider;type:varchar(64);not null"`
	ExternalSubject string     `gorm:"column:external_subject;type:text;not null"`
	AccountID       uuid.UUID  `gorm:"column:account_id;type:uuid;not null;index"`
	Email           *string    `gorm:"column:email;type:varchar(320)"`
	EmailVerified   bool       `gorm:"column:email_verified;not null"`
	DisplayName     *string    `gorm:"column:display_name;type:varchar(255)"`
	ClaimsJSON      string     `gorm:"column:claims_json;type:jsonb;not null"`
	LastLoginAt     *time.Time `gorm:"column:last_login_at;type:timestamptz"`
	CreatedAt       time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt       *time.Time `gorm:"column:updated_at;type:timestamptz"`
}

func (ExternalIdentity) TableName() string { return "auth.external_identities" }
