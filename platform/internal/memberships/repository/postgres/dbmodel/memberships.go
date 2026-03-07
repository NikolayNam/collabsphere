package dbmodel

import (
    "time"

    "github.com/google/uuid"
)

type Membership struct {
    ID             uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
    OrganizationID uuid.UUID `gorm:"column:organization_id;type:uuid;not null;index"`
    AccountID      uuid.UUID `gorm:"column:account_id;type:uuid;not null;index"`
    Role           string    `gorm:"column:role;type:varchar(64);not null"`
    IsActive       bool      `gorm:"column:is_active;not null;default:true"`
    CreatedAt      time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
    UpdatedAt      time.Time `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (Membership) TableName() string { return "iam.memberships" }
