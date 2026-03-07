package http

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/application"
	"github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/mapper"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
)

type Handler struct {
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateOrganization(ctx context.Context, input *dto.CreateOrganizationInput) (*dto.OrganizationResponse, error) {
	principal := authmw.PrincipalFromContext(ctx)
	if !principal.Authenticated {
		return nil, humaerr.From(ctx, fault.Unauthorized("Authentication required"))
	}

	ownerAccountID, err := accdomain.AccountIDFromUUID(principal.AccountID)
	if err != nil {
		return nil, humaerr.From(ctx, fault.Unauthorized("Authentication required"))
	}

	organization, err := h.svc.CreateOrganization(ctx, application.CreateOrganizationCmd{
		Name:           input.Body.Name,
		Slug:           input.Body.Slug,
		OwnerAccountID: ownerAccountID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	return mapper.ToDBOrganizationResponse(organization), nil
}

func (h *Handler) GetOrganizationById(ctx context.Context, input *dto.GetOrganizationByIdInput) (*dto.OrganizationResponse, error) {
	organization, err := h.svc.GetOrganizationById(ctx, application.GetOrganizationByIdQuery{ID: input.ID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	resp := mapper.ToDBOrganizationResponse(organization)
	if resp != nil {
		resp.Status = 200
	}
	return resp, nil
}
