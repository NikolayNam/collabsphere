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
	var dn *string
	if v := t.DisplayName(); v != nil {
		dn = v
	}

	return &dto.OrganizationResponse{
		Status: http.StatusCreated,
		Body: dto.OrganizationBody{
			ID:           t.ID().UUID(),
			PrimaryEmail: t.PrimaryEmail().String(),
			LegalName:    t.LegalName(),
			DisplayName:  dn,
			Status:       string(t.Status()),
		},
	}
}
