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
	Description  *string    `gorm:"column:description;type:text"`
	Website      *string    `gorm:"column:website;type:varchar(512)"`
	PrimaryEmail *string    `gorm:"column:primary_email;type:varchar(320)"`
	Phone        *string    `gorm:"column:phone;type:varchar(32)"`
	Address      *string    `gorm:"column:address;type:text"`
	Industry     *string    `gorm:"column:industry;type:varchar(128)"`
	IsActive     bool       `gorm:"column:is_active;not null;default:true"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (Organization) TableName() string { return "org.organizations" }
