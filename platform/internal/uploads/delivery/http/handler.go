package http

import (
	"context"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/httpbind"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	uploadapp "github.com/NikolayNam/collabsphere/internal/uploads/application"
	"github.com/NikolayNam/collabsphere/internal/uploads/delivery/http/dto"
	uploaddomain "github.com/NikolayNam/collabsphere/internal/uploads/domain"
)

type Handler struct {
	svc *uploadapp.Service
}

func NewHandler(svc *uploadapp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetUpload(ctx context.Context, input *dto.GetUploadInput) (*dto.UploadResponse, error) {
	uploadID, err := httpbind.ParseUUID(input.ID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	upload, err := h.svc.GetUpload(ctx, uploadapp.GetUploadQuery{UploadID: uploadID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return uploadResponse(upload, nil, 200), nil
}

func principal(ctx context.Context) authdomain.Principal {
	return httpbind.Principal(ctx)
}

func uploadResponse(upload *uploaddomain.Upload, uploadURL *string, status int) *dto.UploadResponse {
	out := &dto.UploadResponse{Status: status}
	out.Body.ID = upload.ID
	out.Body.OrganizationID = upload.OrganizationID
	out.Body.ObjectID = upload.ObjectID
	out.Body.CreatedByAccountID = upload.CreatedByAccountID
	out.Body.Purpose = string(upload.Purpose)
	out.Body.Status = string(upload.Status)
	out.Body.Bucket = upload.Bucket
	out.Body.ObjectKey = upload.ObjectKey
	out.Body.FileName = upload.FileName
	out.Body.ContentType = upload.ContentType
	out.Body.DeclaredSizeBytes = upload.DeclaredSizeBytes
	out.Body.ActualSizeBytes = upload.ActualSizeBytes
	out.Body.ChecksumSHA256 = upload.ChecksumSHA256
	out.Body.Metadata = copyMap(upload.Metadata)
	out.Body.ErrorCode = upload.ErrorCode
	out.Body.ErrorMessage = upload.ErrorMessage
	if upload.ResultKind != nil {
		value := string(*upload.ResultKind)
		out.Body.ResultKind = &value
	}
	out.Body.ResultID = upload.ResultID
	out.Body.CompletedAt = upload.CompletedAt
	out.Body.ExpiresAt = upload.ExpiresAt
	out.Body.CreatedAt = upload.CreatedAt
	out.Body.UpdatedAt = upload.UpdatedAt
	out.Body.UploadURL = uploadURL
	return out
}

func copyMap(value map[string]any) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(value))
	for key, item := range value {
		out[key] = item
	}
	return out
}
