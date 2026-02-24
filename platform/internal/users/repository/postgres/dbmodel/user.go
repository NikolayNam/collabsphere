package dbmodel

import (
	basemodel "github.com/NikolayNam/collabsphere-go/shared/contracts/persistence/dbmodel"
)

type User struct {
	basemodel.BaseModel

	Email        string `gorm:"column:email;type:varchar(254);not null"`
	PasswordHash string `gorm:"column:password_hash;type:text;not null"`

	FirstName string  `gorm:"column:first_name;type:text;not null"`
	LastName  string  `gorm:"column:last_name;type:text;not null"`
	Phone     *string `gorm:"column:phone;type:varchar(16)"`

	IsActive bool `gorm:"column:is_active;not null;default:true"`
}

func (User) TableName() string { return "users" }
