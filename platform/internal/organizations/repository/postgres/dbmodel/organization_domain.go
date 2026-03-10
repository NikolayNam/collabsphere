package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationDomain struct {
	ID             uuid.UUID  `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID uuid.UUID  `gorm:"column:organization_id;type:uuid;not null"`
	Hostname       string     `gorm:"column:hostname;type:varchar(253);not null"`
	Kind           string     `gorm:"column:kind;type:varchar(32);not null"`
	IsPrimary      bool       `gorm:"column:is_primary;not null;default:false"`
	VerifiedAt     *time.Time `gorm:"column:verified_at;type:timestamptz"`
	DisabledAt     *time.Time `gorm:"column:disabled_at;type:timestamptz"`
	CreatedAt      time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (OrganizationDomain) TableName() string { return "org.organization_domains" }
