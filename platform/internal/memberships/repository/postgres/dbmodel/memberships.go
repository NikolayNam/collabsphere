package dbmodel

import (
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/persistence/gorm/model"
	"github.com/google/uuid"
)

type Membership struct {
	model.UUIDPK
	model.Timestamps
	model.Blame

	OrganizationID uuid.UUID `gorm:"column:organization_id;type:uuid;not null;index"`
	AccountID      uuid.UUID `gorm:"column:account_id;type:uuid;not null;index"`

	Kind   string `gorm:"column:kind;type:text;not null;default:member"`
	Status string `gorm:"column:status;type:text;not null;default:active"`
}

func (Membership) TableName() string { return "iam.memberships" }

