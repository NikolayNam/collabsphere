package dbmodel

import "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/persistence/gorm/model"

type Organization struct {
	model.UUIDPK
	model.Timestamps

	LegalName    string  `gorm:"column:legal_name;type:text;not null"`
	DisplayName  *string `gorm:"column:display_name;type:text;null"`
	PrimaryEmail string  `gorm:"column:primary_email;type:varchar(254);not null"`

	Status string `gorm:"column:status;not null;default:active"`
}

func (Organization) TableName() string { return "org.organizations" }

