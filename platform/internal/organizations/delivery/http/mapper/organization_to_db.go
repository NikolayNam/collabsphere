package mapper

import (
    "net/http"

    "github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/dto"
    "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

func ToDBOrganizationResponse(t *domain.Organization) *dto.OrganizationResponse {
    if t == nil {
        return nil
    }

    return &dto.OrganizationResponse{
        Status: http.StatusCreated,
        Body: dto.OrganizationBody{
            ID:       t.ID().UUID(),
            Name:     t.Name(),
            Slug:     t.Slug(),
            IsActive: t.IsActive(),
        },
    }
}
