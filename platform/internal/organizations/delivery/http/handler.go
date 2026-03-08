package http

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/application"
	"github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/mapper"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
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
	ownerAccountID, err := principalOrganizationActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	organization, err := h.svc.CreateOrganization(ctx, application.CreateOrganizationCmd{
		Name:           input.Body.Name,
		Slug:           input.Body.Slug,
		OwnerAccountID: ownerAccountID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	return mapper.ToOrganizationResponse(organization, http.StatusCreated), nil
}

func (h *Handler) GetOrganizationById(ctx context.Context, input *dto.GetOrganizationByIdInput) (*dto.OrganizationResponse, error) {
	organization, err := h.svc.GetOrganizationById(ctx, application.GetOrganizationByIdQuery{ID: input.ID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToOrganizationResponse(organization, http.StatusOK), nil
}

func (h *Handler) UpdateOrganization(ctx context.Context, input *dto.UpdateOrganizationInput) (*dto.OrganizationResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organization, err := h.svc.UpdateOrganizationProfile(ctx, application.UpdateOrganizationProfileCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		Name:           input.Body.Name,
		Slug:           input.Body.Slug,
		LogoObjectID:   input.Body.LogoObjectID,
		ClearLogo:      input.Body.ClearLogo,
		Description:    input.Body.Description,
		Website:        input.Body.Website,
		PrimaryEmail:   input.Body.PrimaryEmail,
		Phone:          input.Body.Phone,
		Address:        input.Body.Address,
		Industry:       input.Body.Industry,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToOrganizationResponse(organization, http.StatusOK), nil
}

func (h *Handler) CreateOrganizationLogoUpload(ctx context.Context, input *dto.CreateOrganizationLogoUploadInput) (*dto.UploadResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	result, err := h.svc.CreateOrganizationLogoUpload(ctx, application.CreateOrganizationLogoUploadCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		FileName:       input.Body.FileName,
		ContentType:    input.Body.ContentType,
		SizeBytes:      input.Body.SizeBytes,
		ChecksumSHA256: input.Body.ChecksumSHA256,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	resp := &dto.UploadResponse{Status: http.StatusCreated}
	resp.Body.ObjectID = result.ObjectID
	resp.Body.Bucket = result.Bucket
	resp.Body.ObjectKey = result.ObjectKey
	resp.Body.UploadURL = result.UploadURL
	resp.Body.ExpiresAt = result.ExpiresAt
	resp.Body.FileName = result.FileName
	resp.Body.SizeBytes = result.SizeBytes
	return resp, nil
}

func principalOrganizationActor(ctx context.Context) (accdomain.AccountID, error) {
	principal := authmw.PrincipalFromContext(ctx)
	if !principal.IsAccount() {
		return accdomain.AccountID{}, fault.Unauthorized("Authentication required")
	}
	return accdomain.AccountIDFromUUID(principal.AccountID)
}

func principalOrganizationActorUUID(ctx context.Context) (uuid.UUID, error) {
	principal := authmw.PrincipalFromContext(ctx)
	if !principal.IsAccount() {
		return uuid.Nil, fault.Unauthorized("Authentication required")
	}
	return principal.AccountID, nil
}

func parseOrganizationID(raw string) (orgdomain.OrganizationID, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return orgdomain.OrganizationID{}, fault.Validation("Invalid organization ID")
	}
	return orgdomain.OrganizationIDFromUUID(parsed)
}
