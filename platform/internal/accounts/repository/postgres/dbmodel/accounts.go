package dbmodel

import "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/persistence/gorm/model"

type Account struct {
	model.UUIDPK
	model.Timestamps

	Email        string `gorm:"column:email;type:varchar(254);not null"`
	PasswordHash string `gorm:"column:password_hash;type:text;not null"`

	FirstName string `gorm:"column:first_name;type:varchar(200);not null"`
	LastName  string `gorm:"column:last_name;type:varchar(200);not null"`

	Status string `gorm:"column:status;not null;default:active"`
}

func (Account) TableName() string { return "accounts" }
