package mapper

import (
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/dbmodel"
)

func ToDomainOrganization(m *dbmodel.Organization) (*domain.Organization, error) {
	if m == nil {
		return nil, nil
	}

	id, err := domain.OrganizationIDFromUUID(m.ID)
	if err != nil {
		return nil, err
	}

	email, err := domain.NewEmail(m.PrimaryEmail)
	if err != nil {
		return nil, err
	}

	status, err := domain.NewOrganizationStatus(m.Status)
	if err != nil {
		return nil, err
	}

	return domain.RehydrateOrganization(domain.RehydrateOrganizationParams{
		ID:           id,
		LegalName:    m.LegalName,
		DisplayName:  m.DisplayName,
		PrimaryEmail: email,
		Status:       status,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	})
}
