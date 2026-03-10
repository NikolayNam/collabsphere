package http

import (
	"context"
	"net/http"

	catalogapp "github.com/NikolayNam/collabsphere/internal/catalog/application"
	"github.com/NikolayNam/collabsphere/internal/catalog/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/catalog/delivery/http/mapper"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	uploadsdto "github.com/NikolayNam/collabsphere/internal/uploads/delivery/http/dto"
	"github.com/google/uuid"
)

func (h *Handler) CreateProductImportUpload(ctx context.Context, input *dto.CreateProductImportUploadInput) (*uploadsdto.UploadResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}
	result, err := h.svc.CreateProductImportUpload(ctx, catalogapp.CreateProductImportUploadCmd{
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
	return productImportUploadResponse(result, organizationID.UUID(), actorID.UUID(), input.Body.ContentType, input.Body.ChecksumSHA256), nil
}

func (h *Handler) CompleteProductImportUpload(ctx context.Context, input *dto.CompleteProductImportUploadInput) (*dto.ProductImportResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}
	uploadID, err := parseUUID(input.UploadID)
	if err != nil {
		return nil, humaerr.From(ctx, catalogapp.ErrValidation)
	}
	view, err := h.svc.CompleteProductImportUpload(ctx, catalogapp.CompleteProductImportUploadCmd{
		OrganizationID: organizationID,
		ActorAccountID: actorID,
		UploadID:       uploadID,
		Mode:           input.Body.Mode,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToProductImportResponse(view, http.StatusOK), nil
}

func productImportUploadResponse(result *catalogapp.CreateProductImportUploadResult, organizationID, actorID uuid.UUID, contentType, checksum *string) *uploadsdto.UploadResponse {
	resp := &uploadsdto.UploadResponse{Status: http.StatusCreated}
	resp.Body.ID = result.UploadID
	resp.Body.OrganizationID = &organizationID
	resp.Body.ObjectID = result.ObjectID
	resp.Body.CreatedByAccountID = actorID
	resp.Body.Purpose = "product_import"
	resp.Body.Status = "pending"
	resp.Body.Bucket = result.Bucket
	resp.Body.ObjectKey = result.ObjectKey
	resp.Body.FileName = result.FileName
	resp.Body.ContentType = contentType
	resp.Body.DeclaredSizeBytes = result.SizeBytes
	resp.Body.ChecksumSHA256 = checksum
	resp.Body.ExpiresAt = &result.ExpiresAt
	resp.Body.UploadURL = &result.UploadURL
	return resp
}
