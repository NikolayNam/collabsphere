package http

import (
	"context"
	"net/http"

	"github.com/NikolayNam/collabsphere/internal/organizations/application"
	orgports "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/httpbind"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
)

func (h *Handler) GetOrganizationKYCProfile(ctx context.Context, input *dto.GetOrganizationKYCInput) (*dto.OrganizationKYCResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	profile, docs, err := h.svc.GetOrganizationKYCProfile(ctx, application.GetOrganizationKYCProfileQuery{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.OrganizationKYCResponse{Status: http.StatusOK}
	out.Body.OrganizationID = profile.OrganizationID
	out.Body.Status = profile.Status
	out.Body.LegalName = profile.LegalName
	out.Body.CountryCode = profile.CountryCode
	out.Body.RegistrationNumber = profile.RegistrationNumber
	out.Body.TaxID = profile.TaxID
	out.Body.ReviewNote = profile.ReviewNote
	out.Body.ReviewerAccountID = profile.ReviewerAccountID
	out.Body.SubmittedAt = profile.SubmittedAt
	out.Body.ReviewedAt = profile.ReviewedAt
	out.Body.CreatedAt = profile.CreatedAt
	out.Body.UpdatedAt = profile.UpdatedAt
	out.Body.Documents = make([]dto.OrganizationKYCDocument, 0, len(docs))
	for _, item := range docs {
		out.Body.Documents = append(out.Body.Documents, toOrganizationKYCDocumentDTO(item))
	}
	return out, nil
}

func (h *Handler) UpdateOrganizationKYCProfile(ctx context.Context, input *dto.UpdateOrganizationKYCInput) (*dto.OrganizationKYCResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	profile, err := h.svc.UpdateOrganizationKYCProfile(ctx, application.UpdateOrganizationKYCProfileCmd{
		OrganizationID:     organizationID,
		ActorAccountID:     actorID,
		Status:             input.Body.Status,
		LegalName:          input.Body.LegalName,
		CountryCode:        input.Body.CountryCode,
		RegistrationNumber: input.Body.RegistrationNumber,
		TaxID:              input.Body.TaxID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.OrganizationKYCResponse{Status: http.StatusOK}
	out.Body.OrganizationID = profile.OrganizationID
	out.Body.Status = profile.Status
	out.Body.LegalName = profile.LegalName
	out.Body.CountryCode = profile.CountryCode
	out.Body.RegistrationNumber = profile.RegistrationNumber
	out.Body.TaxID = profile.TaxID
	out.Body.ReviewNote = profile.ReviewNote
	out.Body.ReviewerAccountID = profile.ReviewerAccountID
	out.Body.SubmittedAt = profile.SubmittedAt
	out.Body.ReviewedAt = profile.ReviewedAt
	out.Body.CreatedAt = profile.CreatedAt
	out.Body.UpdatedAt = profile.UpdatedAt
	out.Body.Documents = []dto.OrganizationKYCDocument{}
	return out, nil
}

func (h *Handler) CreateOrganizationKYCDocumentUpload(ctx context.Context, input *dto.CreateOrganizationKYCDocumentUploadInput) (*dto.OrganizationKYCUploadResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	result, err := h.svc.CreateOrganizationKYCDocumentUpload(ctx, application.CreateOrganizationKYCDocumentUploadCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		DocumentType:   input.Body.DocumentType,
		Title:          input.Body.Title,
		FileName:       input.Body.FileName,
		ContentType:    input.Body.ContentType,
		SizeBytes:      input.Body.SizeBytes,
		ChecksumSHA256: input.Body.ChecksumSHA256,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.OrganizationKYCUploadResponse{Status: http.StatusCreated}
	out.Body.ID = result.UploadID
	orgID := organizationID.UUID()
	out.Body.OrganizationID = &orgID
	out.Body.ObjectID = result.ObjectID
	out.Body.CreatedByAccountID = actorID
	out.Body.Purpose = "organization_kyc_document"
	out.Body.Status = "pending"
	out.Body.Bucket = result.Bucket
	out.Body.ObjectKey = result.ObjectKey
	out.Body.FileName = result.FileName
	out.Body.ContentType = input.Body.ContentType
	out.Body.DeclaredSizeBytes = result.SizeBytes
	out.Body.ChecksumSHA256 = input.Body.ChecksumSHA256
	out.Body.Metadata = map[string]any{
		"documentType": result.DocumentType,
		"title":        result.Title,
	}
	out.Body.UploadURL = &result.UploadURL
	out.Body.ExpiresAt = &result.ExpiresAt
	return out, nil
}

func (h *Handler) CompleteOrganizationKYCDocumentUpload(ctx context.Context, input *dto.CompleteOrganizationKYCDocumentUploadInput) (*dto.OrganizationKYCDocumentResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	uploadID, err := httpbind.ParseUUID(input.UploadID, fault.Validation("Upload id is invalid", fault.Field("upload_id", "must be a UUID")))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	item, err := h.svc.CompleteOrganizationKYCDocumentUpload(ctx, application.CompleteOrganizationKYCDocumentUploadCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		UploadID:       uploadID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.OrganizationKYCDocumentResponse{
		Status: http.StatusOK,
		Body:   toOrganizationKYCDocumentDTO(*item),
	}, nil
}

func toOrganizationKYCDocumentDTO(item orgports.OrganizationKYCDocumentRecord) dto.OrganizationKYCDocument {
	return dto.OrganizationKYCDocument{
		ID:                item.ID,
		OrganizationID:    item.OrganizationID,
		ObjectID:          item.ObjectID,
		DocumentType:      item.DocumentType,
		Title:             item.Title,
		Status:            item.Status,
		ReviewNote:        item.ReviewNote,
		ReviewerAccountID: item.ReviewerAccountID,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
		ReviewedAt:        item.ReviewedAt,
	}
}
