package mapper

import (
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/persistence/gorm/model"

	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/dbmodel"
)

func ToDBOrganizationForCreate(t *domain.Organization) *dbmodel.Organization {
	if t == nil {
		return nil
	}

	return &dbmodel.Organization{
		UUIDPK: model.UUIDPK{
			ID: t.ID().UUID(),
		},
		Timestamps: model.Timestamps{
			CreatedAt: t.CreatedAt(),
			UpdatedAt: t.UpdatedAt(),
		},
		LegalName:    t.LegalName(),
		DisplayName:  t.DisplayName(),
		PrimaryEmail: t.PrimaryEmail().String(),
		Status:       string(t.Status()),
	}
}
