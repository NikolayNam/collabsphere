package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type OneTimeCode struct {
	ID           uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	Purpose      string     `gorm:"column:purpose;type:varchar(64);not null"`
	CodeHash     string     `gorm:"column:code_hash;type:text;not null;uniqueIndex"`
	AccountID    uuid.UUID  `gorm:"column:account_id;type:uuid;not null"`
	Provider     string     `gorm:"column:provider;type:varchar(64);not null"`
	Intent       string     `gorm:"column:intent;type:varchar(32);not null"`
	IsNewAccount bool       `gorm:"column:is_new_account;type:boolean;not null"`
	ExpiresAt    time.Time  `gorm:"column:expires_at;type:timestamptz;not null;index"`
	UsedAt       *time.Time `gorm:"column:used_at;type:timestamptz"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;type:timestamptz"`
}

func (OneTimeCode) TableName() string { return "auth.one_time_codes" }
