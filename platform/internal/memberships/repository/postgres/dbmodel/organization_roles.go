package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationRole struct {
	ID             uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID uuid.UUID  `gorm:"column:organization_id;type:uuid;not null;index"`
	Code           string     `gorm:"column:code;type:varchar(64);not null"`
	Name           string     `gorm:"column:name;type:varchar(255);not null"`
	Description    string     `gorm:"column:description;type:text"`
	BaseRole       string     `gorm:"column:base_role;type:varchar(64);not null"`
	CreatedAt      time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;type:timestamptz"`
}

func (OrganizationRole) TableName() string { return "org.organization_roles" }
