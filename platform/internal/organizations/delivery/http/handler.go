package http

import (
    "context"

    "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"

    "github.com/NikolayNam/collabsphere/internal/organizations/application"
    "github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/dto"
    "github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/mapper"
)

type Handler struct {
    svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) CreateOrganization(ctx context.Context, input *dto.CreateOrganizationInput) (*dto.OrganizationResponse, error) {
    t, err := h.svc.CreateOrganization(ctx, application.CreateOrganizationCmd{
        Name: input.Body.Name,
        Slug: input.Body.Slug,
    })
    if err != nil {
        return nil, humaerr.From(ctx, err)
    }

    resp := mapper.ToDBOrganizationResponse(t)
    return resp, nil
}

func (h *Handler) GetOrganizationById(ctx context.Context, input *dto.GetOrganizationByIdInput) (*dto.OrganizationResponse, error) {
    t, err := h.svc.GetOrganizationById(ctx, application.GetOrganizationByIdQuery{ID: input.ID})
    if err != nil {
        return nil, humaerr.From(ctx, err)
    }

    resp := mapper.ToDBOrganizationResponse(t)
    if resp != nil {
        resp.Status = 200
    }
    return resp, nil
}
