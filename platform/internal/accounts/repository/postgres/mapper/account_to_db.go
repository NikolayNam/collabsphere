package mapper

import (
	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres/dbmodel"

	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/persistence/gorm/model"
)

func ToDBAccountForCreate(a *domain.Account) *dbmodel.Account {
	if a == nil {
		return nil
	}

	return &dbmodel.Account{
		UUIDPK: model.UUIDPK{
			ID: a.ID().UUID(),
		},
		Timestamps: model.Timestamps{
			CreatedAt: a.CreatedAt(),
			UpdatedAt: a.UpdatedAt(),
		},
		Email:        a.Email().String(),
		FirstName:    a.FirstName(),
		LastName:     a.LastName(),
		PasswordHash: a.PasswordHash().String(),
		Status:       string(a.Status()),
	}
}
