package mapper

import (
	"github.com/NikolayNam/collabsphere/internal/organizations/application"
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

func ToMyOrganizationsResponse(items []application.MyOrganizationView, status int) *dto.MyOrganizationsResponse {
	resp := &dto.MyOrganizationsResponse{Status: status}
	if len(items) == 0 {
		return resp
	}
	resp.Body.Data = make([]dto.MyOrganizationBody, 0, len(items))
	for _, item := range items {
		resp.Body.Data = append(resp.Body.Data, dto.MyOrganizationBody{
			ID:             item.ID,
			Name:           item.Name,
			Slug:           item.Slug,
			LogoObjectID:   item.LogoObjectID,
			IsActive:       item.IsActive,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
			MembershipRole: item.MembershipRole,
		})
	}
	return resp
}
