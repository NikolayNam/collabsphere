package mapper

import (
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/dbmodel"
)

func ToDomainOrganizationDomain(m *dbmodel.OrganizationDomain) (*domain.OrganizationDomain, error) {
	if m == nil {
		return nil, nil
	}
	organizationID, err := domain.OrganizationIDFromUUID(m.OrganizationID)
	if err != nil {
		return nil, err
	}
	return domain.RehydrateOrganizationDomain(domain.RehydrateOrganizationDomainParams{
		ID:             m.ID,
		OrganizationID: organizationID,
		Hostname:       m.Hostname,
		Kind:           domain.ParseOrganizationDomainKind(m.Kind),
		IsPrimary:      m.IsPrimary,
		VerifiedAt:     m.VerifiedAt,
		DisabledAt:     m.DisabledAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	})
}

func ToDBOrganizationDomainForCreate(d domain.OrganizationDomain) *dbmodel.OrganizationDomain {
	updated := d.CreatedAt()
	if updatedAt := d.UpdatedAt(); updatedAt != nil {
		updated = *updatedAt
	}
	return &dbmodel.OrganizationDomain{
		ID:             d.ID(),
		OrganizationID: d.OrganizationID().UUID(),
		Hostname:       d.Hostname(),
		Kind:           string(d.Kind()),
		IsPrimary:      d.IsPrimary(),
		VerifiedAt:     d.VerifiedAt(),
		DisabledAt:     d.DisabledAt(),
		CreatedAt:      d.CreatedAt(),
		UpdatedAt:      updated,
	}
}
