package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type Group struct {
	ID          uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	Name        string     `gorm:"column:name;type:varchar(255);not null"`
	Slug        string     `gorm:"column:slug;type:varchar(255);not null"`
	Description *string    `gorm:"column:description;type:text"`
	IsActive    bool       `gorm:"column:is_active;not null"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;type:timestamptz"`
	CreatedBy   *uuid.UUID `gorm:"column:created_by;type:uuid"`
	UpdatedBy   *uuid.UUID `gorm:"column:updated_by;type:uuid"`
}

func (Group) TableName() string { return "iam.groups" }
