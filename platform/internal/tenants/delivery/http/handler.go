package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	tenantsapp "github.com/NikolayNam/collabsphere/internal/tenants/application"
	"github.com/NikolayNam/collabsphere/internal/tenants/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/tenants/domain"
	"github.com/google/uuid"
)

type Handler struct {
	svc *tenantsapp.Service
}

func NewHandler(svc *tenantsapp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateTenant(ctx context.Context, input *dto.CreateTenantInput) (*dto.TenantResponse, error) {
	actorID, err := principalTenantActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	tenant, err := h.svc.CreateTenant(ctx, actorID, input.Body.Name, input.Body.Slug, input.Body.Description)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.TenantResponse{
		Status: http.StatusCreated,
		Body:   toTenantPayload(*tenant),
	}, nil
}

func (h *Handler) ListMyTenants(ctx context.Context, input *dto.ListMyTenantsInput) (*dto.MyTenantsResponse, error) {
	actorID, err := principalTenantActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	items, err := h.svc.ListMyTenants(ctx, actorID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.MyTenantsResponse{Status: http.StatusOK}
	out.Body.Data = make([]dto.MyTenantPayload, 0, len(items))
	for _, item := range items {
		out.Body.Data = append(out.Body.Data, dto.MyTenantPayload{
			ID:             item.ID,
			Name:           item.Name,
			Slug:           item.Slug,
			Description:    item.Description,
			IsActive:       item.IsActive,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
			MembershipRole: string(item.MembershipRole),
		})
	}
	return out, nil
}

func (h *Handler) GetTenant(ctx context.Context, input *dto.GetTenantInput) (*dto.TenantResponse, error) {
	actorID, err := principalTenantActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	tenantID, err := parseUUID(input.ID, "Invalid tenant id")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	tenant, err := h.svc.GetTenant(ctx, actorID, tenantID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.TenantResponse{
		Status: http.StatusOK,
		Body:   toTenantPayload(*tenant),
	}, nil
}

func (h *Handler) AddTenantMember(ctx context.Context, input *dto.AddTenantMemberInput) (*dto.EmptyResponse, error) {
	actorID, err := principalTenantActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	tenantID, err := parseUUID(input.TenantID, "Invalid tenant id")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	accountID, err := parseUUID(input.Body.AccountID, "Invalid account id")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	if err := h.svc.AddTenantMember(ctx, actorID, tenantID, accountID, strings.TrimSpace(input.Body.Role)); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.EmptyResponse{Status: http.StatusCreated}, nil
}

func (h *Handler) ListTenantMembers(ctx context.Context, input *dto.ListTenantMembersInput) (*dto.TenantMembersListResponse, error) {
	actorID, err := principalTenantActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	tenantID, err := parseUUID(input.TenantID, "Invalid tenant id")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	items, err := h.svc.ListTenantMembers(ctx, actorID, tenantID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.TenantMembersListResponse{Status: http.StatusOK}
	out.Body.Members = make([]dto.TenantMemberPayload, 0, len(items))
	for _, item := range items {
		out.Body.Members = append(out.Body.Members, dto.TenantMemberPayload{
			ID:        item.ID,
			TenantID:  item.TenantID,
			AccountID: item.AccountID,
			Role:      string(item.Role),
			IsActive:  item.IsActive,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			DeletedAt: item.DeletedAt,
		})
	}
	return out, nil
}

func (h *Handler) AddTenantOrganization(ctx context.Context, input *dto.AddTenantOrganizationInput) (*dto.EmptyResponse, error) {
	actorID, err := principalTenantActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	tenantID, err := parseUUID(input.TenantID, "Invalid tenant id")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseUUID(input.Body.OrganizationID, "Invalid organization id")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	if err := h.svc.AddTenantOrganization(ctx, actorID, tenantID, organizationID); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.EmptyResponse{Status: http.StatusCreated}, nil
}

func (h *Handler) ListTenantOrganizations(ctx context.Context, input *dto.ListTenantOrganizationsInput) (*dto.TenantOrganizationsListResponse, error) {
	actorID, err := principalTenantActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	tenantID, err := parseUUID(input.TenantID, "Invalid tenant id")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	items, err := h.svc.ListTenantOrganizations(ctx, actorID, tenantID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.TenantOrganizationsListResponse{Status: http.StatusOK}
	out.Body.Organizations = make([]dto.TenantOrganizationPayload, 0, len(items))
	for _, item := range items {
		out.Body.Organizations = append(out.Body.Organizations, dto.TenantOrganizationPayload{
			ID:             item.ID,
			TenantID:       item.TenantID,
			OrganizationID: item.OrganizationID,
			IsActive:       item.IsActive,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
			DeletedAt:      item.DeletedAt,
		})
	}
	return out, nil
}

func principalTenantActor(ctx context.Context) (uuid.UUID, error) {
	principal := authmw.PrincipalFromContext(ctx)
	if !principal.IsAccount() {
		return uuid.Nil, fault.Unauthorized("Authentication required")
	}
	return principal.AccountID, nil
}

func parseUUID(raw string, message string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fault.Validation(message)
	}
	return id, nil
}

func toTenantPayload(tenant domain.Tenant) dto.TenantPayload {
	return dto.TenantPayload{
		ID:          tenant.ID,
		Name:        tenant.Name,
		Slug:        tenant.Slug,
		Description: tenant.Description,
		IsActive:    tenant.IsActive,
		CreatedAt:   tenant.CreatedAt,
		UpdatedAt:   tenant.UpdatedAt,
	}
}
