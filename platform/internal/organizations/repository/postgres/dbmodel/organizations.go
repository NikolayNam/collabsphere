package dbmodel

import (
    "time"

    "github.com/google/uuid"
)

type Organization struct {
    ID           uuid.UUID  `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
    Name         string     `gorm:"column:name;type:varchar(255);not null"`
    Slug         string     `gorm:"column:slug;type:varchar(255);not null"`
    LogoObjectID *uuid.UUID `gorm:"column:logo_object_id;type:uuid"`
    IsActive     bool       `gorm:"column:is_active;not null;default:true"`
    CreatedAt    time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
    UpdatedAt    time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (Organization) TableName() string { return "org.organizations" }
