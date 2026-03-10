package http

import (
	"context"
	"net/http"

	"github.com/NikolayNam/collabsphere/internal/organizations/application"
	"github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/mapper"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	uploadsdto "github.com/NikolayNam/collabsphere/internal/uploads/delivery/http/dto"
	"github.com/google/uuid"
)

func (h *Handler) CreateOrganizationLegalDocumentUpload(ctx context.Context, input *dto.CreateOrganizationLegalDocumentUploadInput) (*uploadsdto.UploadResponse, error) {
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
		Title:          derefString(input.Body.Title),
		FileName:       input.Body.FileName,
		ContentType:    input.Body.ContentType,
		SizeBytes:      input.Body.SizeBytes,
		ChecksumSHA256: input.Body.ChecksumSHA256,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return organizationUploadResponse(result, organizationID.UUID(), actorID, input.Body.ContentType, input.Body.ChecksumSHA256), nil
}

func (h *Handler) CompleteOrganizationLegalDocumentUpload(ctx context.Context, input *dto.CompleteOrganizationLegalDocumentUploadInput) (*dto.OrganizationLegalDocumentResponse, error) {
	actorID, err := principalOrganizationActorUUID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	uploadID, err := parseUUID(input.UploadID, "Invalid upload ID")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	document, err := h.svc.CompleteLegalDocumentUpload(ctx, application.CompleteLegalDocumentUploadCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		UploadID:       uploadID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToOrganizationLegalDocumentResponse(document, http.StatusOK), nil
}

func organizationUploadResponse(result *application.CreateLegalDocumentUploadResult, organizationID, actorID uuid.UUID, contentType, checksum *string) *uploadsdto.UploadResponse {
	resp := &uploadsdto.UploadResponse{Status: http.StatusCreated}
	resp.Body.ID = result.UploadID
	resp.Body.OrganizationID = &organizationID
	resp.Body.ObjectID = result.ObjectID
	resp.Body.CreatedByAccountID = actorID
	resp.Body.Purpose = "organization_legal_document"
	resp.Body.Status = "pending"
	resp.Body.Bucket = result.Bucket
	resp.Body.ObjectKey = result.ObjectKey
	resp.Body.FileName = result.FileName
	resp.Body.ContentType = contentType
	resp.Body.DeclaredSizeBytes = result.SizeBytes
	resp.Body.ChecksumSHA256 = checksum
	resp.Body.Metadata = map[string]any{"documentType": result.DocumentType}
	resp.Body.ExpiresAt = &result.ExpiresAt
	resp.Body.UploadURL = &result.UploadURL
	return resp
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
