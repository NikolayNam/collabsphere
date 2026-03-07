package mapper

import (
    "github.com/NikolayNam/collabsphere/internal/organizations/domain"
    "github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/dbmodel"
)

func ToDBOrganizationForCreate(t *domain.Organization) *dbmodel.Organization {
    if t == nil {
        return nil
    }

    updatedAt := t.CreatedAt()
    if t.UpdatedAt() != nil {
        updatedAt = *t.UpdatedAt()
    }

    return &dbmodel.Organization{
        ID:        t.ID().UUID(),
        Name:      t.Name(),
        Slug:      t.Slug(),
        IsActive:  t.IsActive(),
        CreatedAt: t.CreatedAt(),
        UpdatedAt: updatedAt,
    }
}
