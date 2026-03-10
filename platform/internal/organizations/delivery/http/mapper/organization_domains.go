package mapper

import (
	"github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

func ToOrganizationDomainBodies(items []domain.OrganizationDomain) []dto.OrganizationDomainBody {
	if len(items) == 0 {
		return nil
	}
	out := make([]dto.OrganizationDomainBody, 0, len(items))
	for _, item := range items {
		out = append(out, dto.OrganizationDomainBody{
			ID:         item.ID(),
			Hostname:   item.Hostname(),
			Kind:       string(item.Kind()),
			IsPrimary:  item.IsPrimary(),
			IsVerified: item.IsVerified(),
			VerifiedAt: item.VerifiedAt(),
			CreatedAt:  item.CreatedAt(),
			UpdatedAt:  item.UpdatedAt(),
		})
	}
	return out
}
