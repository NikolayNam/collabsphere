package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID             uuid.UUID  `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Email          string     `gorm:"column:email;type:varchar(320);not null"`
	DisplayName    *string    `gorm:"column:display_name;type:varchar(255)"`
	AvatarObjectID *uuid.UUID `gorm:"column:avatar_object_id;type:uuid"`
	Bio            *string    `gorm:"column:bio;type:text"`
	Phone          *string    `gorm:"column:phone;type:varchar(32)"`
	Locale         *string    `gorm:"column:locale;type:varchar(16)"`
	Timezone       *string    `gorm:"column:timezone;type:varchar(64)"`
	Website        *string    `gorm:"column:website;type:varchar(512)"`
	IsActive       bool       `gorm:"column:is_active;not null;default:true"`
	CreatedAt      time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (Account) TableName() string { return "iam.accounts" }
