package mapper

import (
	"github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

func ToOrganizationResponse(t *domain.Organization, status int) *dto.OrganizationResponse {
	if t == nil {
		return nil
	}

	return &dto.OrganizationResponse{
		Status: status,
		Body: dto.OrganizationBody{
			ID:           t.ID().UUID(),
			Name:         t.Name(),
			Slug:         t.Slug(),
			LogoObjectID: t.LogoObjectID(),
			Description:  t.Description(),
			Website:      t.Website(),
			PrimaryEmail: t.PrimaryEmail(),
			Phone:        t.Phone(),
			Address:      t.Address(),
			Industry:     t.Industry(),
			IsActive:     t.IsActive(),
			CreatedAt:    t.CreatedAt(),
			UpdatedAt:    t.UpdatedAt(),
		},
	}
}
