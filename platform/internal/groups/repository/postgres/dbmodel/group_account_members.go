package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type GroupAccountMember struct {
	ID        uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	GroupID   uuid.UUID  `gorm:"column:group_id;type:uuid;not null"`
	AccountID uuid.UUID  `gorm:"column:account_id;type:uuid;not null"`
	Role      string     `gorm:"column:role;type:varchar(64);not null"`
	IsActive  bool       `gorm:"column:is_active;not null"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamptz"`
}

func (GroupAccountMember) TableName() string { return "iam.group_account_members" }
