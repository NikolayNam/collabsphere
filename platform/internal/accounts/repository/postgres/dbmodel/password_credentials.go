package dbmodel

import (
    "time"

    "github.com/google/uuid"
)

type PasswordCredential struct {
    AccountID    uuid.UUID `gorm:"column:account_id;type:uuid;primaryKey"`
    PasswordHash string    `gorm:"column:password_hash;type:text;not null"`
    CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
    UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (PasswordCredential) TableName() string { return "auth.password_credentials" }
