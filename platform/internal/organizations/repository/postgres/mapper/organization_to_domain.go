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

    return domain.RehydrateOrganization(domain.RehydrateOrganizationParams{
        ID:        id,
        Name:      m.Name,
        Slug:      m.Slug,
        IsActive:  m.IsActive,
        CreatedAt: m.CreatedAt,
        UpdatedAt: m.UpdatedAt,
    })
}
