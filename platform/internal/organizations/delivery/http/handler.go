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
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/httpbind"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
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
		Domains:        mapOrganizationDomainDrafts(input.Body.Domains),
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return h.organizationResponse(ctx, organization, http.StatusCreated)
}

func (h *Handler) ListMyOrganizations(ctx context.Context, input *struct{}) (*dto.MyOrganizationsResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	items, err := h.svc.ListMyOrganizations(ctx, actorID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToMyOrganizationsResponse(items, http.StatusOK), nil
}

func (h *Handler) GetOrganizationById(ctx context.Context, input *dto.GetOrganizationByIdInput) (*dto.OrganizationResponse, error) {
	organization, err := h.svc.GetOrganizationById(ctx, application.GetOrganizationByIdQuery{ID: input.ID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return h.organizationResponse(ctx, organization, http.StatusOK)
}

func (h *Handler) ResolveOrganizationByHost(ctx context.Context, input *dto.ResolveOrganizationByHostInput) (*dto.OrganizationResponse, error) {
	organization, err := h.svc.GetOrganizationByHost(ctx, input.Host)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	if organization == nil {
		return nil, humaerr.From(ctx, fault.NotFound("Organization not found", fault.Code("ORGANIZATION_NOT_FOUND")))
	}
	return h.organizationResponse(ctx, organization, http.StatusOK)
}

func (h *Handler) ListPublicKYCDirectory(ctx context.Context, input *dto.ListPublicKYCDirectoryInput) (*dto.PublicKYCDirectoryResponse, error) {
	items, err := h.svc.ListPublicKYCDirectoryOrganizations(ctx, input.Limit)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.PublicKYCDirectoryResponse{Status: http.StatusOK}
	out.Body.Items = make([]dto.PublicKYCDirectoryOrganizationBody, 0, len(items))
	for _, item := range items {
		out.Body.Items = append(out.Body.Items, dto.PublicKYCDirectoryOrganizationBody{
			ID:            item.ID.String(),
			Name:          item.Name,
			Slug:          item.Slug,
			Description:   item.Description,
			Website:       item.Website,
			Industry:      item.Industry,
			PrimaryDomain: item.PrimaryDomain,
			KYCLevelCode:  item.KYCLevelCode,
			KYCLevelName:  item.KYCLevelName,
		})
	}
	return out, nil
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
		Domains:        mapOrganizationDomainDraftsPtr(input.Body.Domains),
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
	return h.organizationResponse(ctx, organization, http.StatusOK)
}

func (h *Handler) UploadOrganizationLogo(ctx context.Context, input *dto.UploadOrganizationLogoInput) (*dto.OrganizationResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	form := input.RawBody.Data()
	if form == nil || !form.File.IsSet {
		return nil, humaerr.From(ctx, fault.Validation("Logo file is required"))
	}
	defer form.File.Close()
	fileName := form.File.Filename
	if fileName == "" {
		fileName = "logo.bin"
	}
	organization, err := h.svc.UploadOrganizationLogo(ctx, application.UploadOrganizationLogoCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		FileName:       fileName,
		ContentType:    form.File.ContentType,
		SizeBytes:      form.File.Size,
		Body:           form.File,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return h.organizationResponse(ctx, organization, http.StatusOK)
}

func (h *Handler) UploadOrganizationVideo(ctx context.Context, input *dto.UploadOrganizationVideoInput) (*dto.OrganizationResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	form := input.RawBody.Data()
	if form == nil || !form.File.IsSet {
		return nil, humaerr.From(ctx, fault.Validation("Organization video file is required"))
	}
	defer form.File.Close()
	fileName := form.File.Filename
	if fileName == "" {
		fileName = "organization-video.mp4"
	}
	if _, err := h.svc.UploadOrganizationVideo(ctx, application.UploadOrganizationVideoCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		FileName:       fileName,
		ContentType:    form.File.ContentType,
		SizeBytes:      form.File.Size,
		Body:           form.File,
	}); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organization, err := h.svc.GetOrganizationById(ctx, application.GetOrganizationByIdQuery{ID: input.ID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return h.organizationResponse(ctx, organization, http.StatusOK)
}

func (h *Handler) ListOrganizationVideos(ctx context.Context, input *dto.ListOrganizationVideosInput) (*dto.OrganizationVideosResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	items, err := h.svc.ListOrganizationVideos(ctx, organizationID, actorID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	resp := &dto.OrganizationVideosResponse{Status: http.StatusOK}
	resp.Body.Items = make([]dto.OrganizationVideoItem, 0, len(items))
	for _, item := range items {
		resp.Body.Items = append(resp.Body.Items, dto.OrganizationVideoItem{
			ID:          item.ID,
			ObjectID:    item.ObjectID,
			FileName:    item.FileName,
			ContentType: item.ContentType,
			SizeBytes:   item.SizeBytes,
			CreatedAt:   item.CreatedAt,
			UploadedBy:  item.UploadedBy,
			SortOrder:   item.SortOrder,
		})
	}
	return resp, nil
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
	applicationView, err := h.svc.GetCooperationApplication(ctx, application.GetCooperationApplicationQuery{OrganizationID: organizationID, ActorAccountID: actorID})
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
	applicationView, err := h.svc.UpdateCooperationApplication(ctx, application.UpdateCooperationApplicationCmd{OrganizationID: organizationID, ActorAccountID: actorID, ConfirmationEmail: input.Body.ConfirmationEmail, CompanyName: input.Body.CompanyName, RepresentedCategories: input.Body.RepresentedCategories, MinimumOrderAmount: input.Body.MinimumOrderAmount, DeliveryGeography: input.Body.DeliveryGeography, SalesChannels: input.Body.SalesChannels, StorefrontURL: input.Body.StorefrontURL, ContactFirstName: input.Body.ContactFirstName, ContactLastName: input.Body.ContactLastName, ContactJobTitle: input.Body.ContactJobTitle, PriceListObjectID: input.Body.PriceListObjectID, PriceListStatus: input.Body.PriceListStatus, ClearPriceList: input.Body.ClearPriceList, ContactEmail: input.Body.ContactEmail, ContactPhone: input.Body.ContactPhone, PartnerCode: input.Body.PartnerCode})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToCooperationApplicationResponse(applicationView, http.StatusOK), nil
}

func (h *Handler) PublishAllCatalog(ctx context.Context, input *dto.PublishAllCatalogInput) (*dto.PublishAllCatalogResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	if err := h.svc.PublishAllCatalog(ctx, application.PublishAllCatalogCmd{OrganizationID: organizationID, ActorAccountID: actorID}); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.PublishAllCatalogResponse{Status: http.StatusNoContent}, nil
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
	applicationView, err := h.svc.SubmitCooperationApplication(ctx, application.SubmitCooperationApplicationCmd{OrganizationID: organizationID, ActorAccountID: actorID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToCooperationApplicationResponse(applicationView, http.StatusOK), nil
}

func (h *Handler) UploadCooperationPriceList(ctx context.Context, input *dto.UploadCooperationPriceListInput) (*dto.CooperationApplicationResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	form := input.RawBody.Data()
	if form == nil || !form.File.IsSet {
		return nil, humaerr.From(ctx, fault.Validation("Price list file is required"))
	}
	defer form.File.Close()
	fileName := form.File.Filename
	if fileName == "" {
		fileName = "price-list.xlsx"
	}
	applicationView, err := h.svc.UploadCooperationPriceList(ctx, application.UploadCooperationPriceListCmd{OrganizationID: organizationID, ActorAccountID: actorID, FileName: fileName, ContentType: form.File.ContentType, SizeBytes: form.File.Size, Body: form.File})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToCooperationApplicationResponse(applicationView, http.StatusOK), nil
}

func (h *Handler) UploadOrganizationLegalDocument(ctx context.Context, input *dto.UploadOrganizationLegalDocumentInput) (*dto.OrganizationLegalDocumentResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	form := input.RawBody.Data()
	if form == nil || !form.File.IsSet {
		return nil, humaerr.From(ctx, fault.Validation("Legal document file is required"))
	}
	defer form.File.Close()
	fileName := form.File.Filename
	if fileName == "" {
		fileName = "document.pdf"
	}
	document, err := h.svc.UploadOrganizationLegalDocument(ctx, application.UploadOrganizationLegalDocumentCmd{OrganizationID: organizationID, ActorAccountID: actorID, DocumentType: form.DocumentType, Title: form.Title, FileName: fileName, ContentType: form.File.ContentType, SizeBytes: form.File.Size, Body: form.File})
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

func (h *Handler) GetOrganizationKYCRequirements(ctx context.Context, input *dto.GetOrganizationKYCRequirementsInput) (*dto.OrganizationKYCRequirementsResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	requirements, err := h.svc.GetOrganizationKYCRequirements(ctx, application.GetOrganizationKYCRequirementsQuery{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToOrganizationKYCRequirementsResponse(requirements, http.StatusOK), nil
}

func (h *Handler) GetOrganizationLegalDocumentAnalysis(ctx context.Context, input *dto.GetOrganizationLegalDocumentAnalysisInput) (*dto.OrganizationLegalDocumentAnalysisResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	documentID, err := parseUUID(input.DocumentID, "Invalid legal document ID")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	analysis, err := h.svc.GetOrganizationLegalDocumentAnalysis(ctx, application.GetOrganizationLegalDocumentAnalysisQuery{OrganizationID: organizationID, ActorAccountID: actorID, DocumentID: documentID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToOrganizationLegalDocumentAnalysisResponse(analysis, http.StatusOK), nil
}

func (h *Handler) GetOrganizationLegalDocumentVerification(ctx context.Context, input *dto.GetOrganizationLegalDocumentVerificationInput) (*dto.OrganizationLegalDocumentVerificationResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	documentID, err := parseUUID(input.DocumentID, "Invalid legal document ID")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	verification, err := h.svc.GetOrganizationLegalDocumentVerification(ctx, application.GetOrganizationLegalDocumentVerificationQuery{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		DocumentID:     documentID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToOrganizationLegalDocumentVerificationResponse(verification, http.StatusOK), nil
}

func (h *Handler) ReprocessOrganizationLegalDocumentAnalysis(ctx context.Context, input *dto.ReprocessOrganizationLegalDocumentAnalysisInput) (*dto.OrganizationLegalDocumentAnalysisResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	documentID, err := parseUUID(input.DocumentID, "Invalid legal document ID")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	analysis, err := h.svc.ReprocessOrganizationLegalDocumentAnalysis(ctx, application.ReprocessOrganizationLegalDocumentAnalysisCmd{OrganizationID: organizationID, ActorAccountID: actorID, DocumentID: documentID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToOrganizationLegalDocumentAnalysisResponse(analysis, http.StatusOK), nil
}

func (h *Handler) organizationResponse(ctx context.Context, t *orgdomain.Organization, status int) (*dto.OrganizationResponse, error) {
	resp := mapper.ToOrganizationResponse(t, status)
	if resp == nil || t == nil {
		return resp, nil
	}
	ids, err := h.svc.ListOrganizationVideoObjectIDs(ctx, t.ID())
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	domains, err := h.svc.ListOrganizationDomains(ctx, t.ID())
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	resp.Body.VideoObjectIDs = ids
	resp.Body.Domains = mapper.ToOrganizationDomainBodies(domains)
	return resp, nil
}

func mapOrganizationDomainDrafts(items []dto.OrganizationDomainInput) []orgdomain.OrganizationDomainDraft {
	if len(items) == 0 {
		return nil
	}
	out := make([]orgdomain.OrganizationDomainDraft, 0, len(items))
	for _, item := range items {
		out = append(out, orgdomain.OrganizationDomainDraft{
			Hostname:  item.Hostname,
			Kind:      item.Kind,
			IsPrimary: item.IsPrimary,
		})
	}
	return out
}

func mapOrganizationDomainDraftsPtr(items *[]dto.OrganizationDomainInput) *[]orgdomain.OrganizationDomainDraft {
	if items == nil {
		return nil
	}
	mapped := mapOrganizationDomainDrafts(*items)
	return &mapped
}

func principalOrganizationActor(ctx context.Context) (accdomain.AccountID, error) {
	return httpbind.RequireAccountID(ctx, fault.Unauthorized("Authentication required"))
}
func principalOrganizationActorUUID(ctx context.Context) (uuid.UUID, error) {
	return httpbind.RequireAccountUUID(ctx, fault.Unauthorized("Authentication required"))
}
func parseOrganizationID(raw string) (orgdomain.OrganizationID, error) {
	return httpbind.ParseOrganizationID(raw, fault.Validation("Invalid organization ID"))
}
func parseUUID(raw, message string) (uuid.UUID, error) {
	return httpbind.ParseUUID(raw, fault.Validation(message))
}
