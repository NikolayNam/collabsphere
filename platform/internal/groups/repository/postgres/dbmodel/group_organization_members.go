package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type GroupOrganizationMember struct {
	ID             uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	GroupID        uuid.UUID  `gorm:"column:group_id;type:uuid;not null"`
	OrganizationID uuid.UUID  `gorm:"column:organization_id;type:uuid;not null"`
	IsActive       bool       `gorm:"column:is_active;not null"`
	CreatedAt      time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;type:timestamptz"`
}

func (GroupOrganizationMember) TableName() string { return "iam.group_organization_members" }
