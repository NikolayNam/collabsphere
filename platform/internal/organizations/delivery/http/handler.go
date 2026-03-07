package http

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/organizations/application"
	"github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/mapper"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
)

type Handler struct {
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) CreateOrganization(ctx context.Context, input *dto.CreateOrganizationInput) (*dto.OrganizationResponse, error) {
	t, err := h.svc.CreateOrganization(ctx, application.CreateOrganizationCmd{
		LegalName:    input.Body.LegalName,
		DisplayName:  input.Body.DisplayName,
		PrimaryEmail: input.Body.PrimaryEmail,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	resp := mapper.ToDBOrganizationResponse(t)
	return resp, nil
}

func (h *Handler) GetOrganizationById(ctx context.Context, input *dto.GetOrganizationByIdInput) (*dto.OrganizationResponse, error) {
	t, err := h.svc.GetOrganizationById(ctx, application.GetOrganizationByIdQuery{
		ID: input.ID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	resp := mapper.ToDBOrganizationResponse(t)
	return resp, nil
}
