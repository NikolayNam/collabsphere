package http

import (
	"context"
	"net/http"
	"time"

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
	return toUploadResponse(result.ObjectID, result.Bucket, result.ObjectKey, result.UploadURL, result.ExpiresAt, result.FileName, result.SizeBytes, http.StatusCreated), nil
}

func (h *Handler) GetCooperationApplication(ctx context.Context, input *dto.GetCooperationApplicationInput) (*dto.CooperationApplicationResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	applicationView, err := h.svc.GetCooperationApplication(ctx, application.GetCooperationApplicationQuery{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToCooperationApplicationResponse(applicationView, http.StatusOK), nil
}

func (h *Handler) UpdateCooperationApplication(ctx context.Context, input *dto.UpdateCooperationApplicationInput) (*dto.CooperationApplicationResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	applicationView, err := h.svc.UpdateCooperationApplication(ctx, application.UpdateCooperationApplicationCmd{
		OrganizationID:        organizationID,
		ActorAccountID:        actorID,
		ConfirmationEmail:     input.Body.ConfirmationEmail,
		CompanyName:           input.Body.CompanyName,
		RepresentedCategories: input.Body.RepresentedCategories,
		MinimumOrderAmount:    input.Body.MinimumOrderAmount,
		DeliveryGeography:     input.Body.DeliveryGeography,
		SalesChannels:         input.Body.SalesChannels,
		StorefrontURL:         input.Body.StorefrontURL,
		ContactFirstName:      input.Body.ContactFirstName,
		ContactLastName:       input.Body.ContactLastName,
		ContactJobTitle:       input.Body.ContactJobTitle,
		PriceListObjectID:     input.Body.PriceListObjectID,
		ClearPriceList:        input.Body.ClearPriceList,
		ContactEmail:          input.Body.ContactEmail,
		ContactPhone:          input.Body.ContactPhone,
		PartnerCode:           input.Body.PartnerCode,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToCooperationApplicationResponse(applicationView, http.StatusOK), nil
}

func (h *Handler) SubmitCooperationApplication(ctx context.Context, input *dto.SubmitCooperationApplicationInput) (*dto.CooperationApplicationResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	applicationView, err := h.svc.SubmitCooperationApplication(ctx, application.SubmitCooperationApplicationCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToCooperationApplicationResponse(applicationView, http.StatusOK), nil
}

func (h *Handler) CreateCooperationPriceListUpload(ctx context.Context, input *dto.CreateCooperationPriceListUploadInput) (*dto.UploadResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	result, err := h.svc.CreateCooperationPriceListUpload(ctx, application.CreateCooperationPriceListUploadCmd{
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
	return toUploadResponse(result.ObjectID, result.Bucket, result.ObjectKey, result.UploadURL, result.ExpiresAt, result.FileName, result.SizeBytes, http.StatusCreated), nil
}

func (h *Handler) CreateOrganizationLegalDocumentUpload(ctx context.Context, input *dto.CreateLegalDocumentUploadInput) (*dto.UploadResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	result, err := h.svc.CreateLegalDocumentUpload(ctx, application.CreateLegalDocumentUploadCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		DocumentType:   input.Body.DocumentType,
		FileName:       input.Body.FileName,
		ContentType:    input.Body.ContentType,
		SizeBytes:      input.Body.SizeBytes,
		ChecksumSHA256: input.Body.ChecksumSHA256,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return toUploadResponse(result.ObjectID, result.Bucket, result.ObjectKey, result.UploadURL, result.ExpiresAt, result.FileName, result.SizeBytes, http.StatusCreated), nil
}

func (h *Handler) CreateOrganizationLegalDocument(ctx context.Context, input *dto.CreateOrganizationLegalDocumentInput) (*dto.OrganizationLegalDocumentResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	document, err := h.svc.AddOrganizationLegalDocument(ctx, application.AddOrganizationLegalDocumentCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		DocumentType:   input.Body.DocumentType,
		ObjectID:       input.Body.ObjectID,
		Title:          input.Body.Title,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToOrganizationLegalDocumentResponse(document, http.StatusCreated), nil
}

func (h *Handler) ListOrganizationLegalDocuments(ctx context.Context, input *dto.ListOrganizationLegalDocumentsInput) (*dto.OrganizationLegalDocumentsResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	documents, err := h.svc.ListOrganizationLegalDocuments(ctx, organizationID, actorID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToOrganizationLegalDocumentsResponse(documents, http.StatusOK), nil
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

func toUploadResponse(objectID uuid.UUID, bucket, objectKey, uploadURL string, expiresAt time.Time, fileName string, sizeBytes int64, status int) *dto.UploadResponse {
	resp := &dto.UploadResponse{Status: status}
	resp.Body.ObjectID = objectID
	resp.Body.Bucket = bucket
	resp.Body.ObjectKey = objectKey
	resp.Body.UploadURL = uploadURL
	resp.Body.ExpiresAt = expiresAt
	resp.Body.FileName = fileName
	resp.Body.SizeBytes = sizeBytes
	return resp
}
