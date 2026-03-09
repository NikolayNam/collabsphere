package http

import (
	"context"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/httpbind"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	storageapp "github.com/NikolayNam/collabsphere/internal/storage/application"
	"github.com/NikolayNam/collabsphere/internal/storage/delivery/http/dto"
)

type Handler struct {
	svc *storageapp.Service
}

func NewHandler(svc *storageapp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) DownloadObject(ctx context.Context, input *dto.DownloadObjectInput) (*dto.DownloadObjectResponse, error) {
	objectID, err := httpbind.ParseUUID(input.ObjectID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	result, err := h.svc.CreateDownload(ctx, storageapp.DownloadObjectQuery{ObjectID: objectID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.DownloadObjectResponse{Status: 200}
	out.Body.ObjectID = result.ObjectID
	out.Body.OrganizationID = result.OrganizationID
	out.Body.FileName = result.FileName
	out.Body.ContentType = result.ContentType
	out.Body.SizeBytes = result.SizeBytes
	out.Body.DownloadURL = result.DownloadURL
	out.Body.ExpiresAt = result.ExpiresAt
	out.Body.CreatedAt = result.CreatedAt
	return out, nil
}

func (h *Handler) ListMyFiles(ctx context.Context, _ *dto.ListMyFilesInput) (*dto.ListFilesResponse, error) {
	files, err := h.svc.ListMyFiles(ctx, storageapp.ListMyFilesQuery{Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return listFilesResponse(files), nil
}

func (h *Handler) ListOrganizationFiles(ctx context.Context, input *dto.ListOrganizationFilesInput) (*dto.ListFilesResponse, error) {
	organizationID, err := httpbind.ParseUUID(input.OrganizationID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	files, err := h.svc.ListOrganizationFiles(ctx, storageapp.ListOrganizationFilesQuery{
		OrganizationID: organizationID,
		Actor:          principal(ctx),
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return listFilesResponse(files), nil
}

func principal(ctx context.Context) authdomain.Principal {
	return httpbind.Principal(ctx)
}

func listFilesResponse(files []storageapp.ListedFile) *dto.ListFilesResponse {
	out := &dto.ListFilesResponse{Status: 200}
	out.Body.Items = make([]dto.FileItem, 0, len(files))
	for _, file := range files {
		out.Body.Items = append(out.Body.Items, dto.FileItem{
			ObjectID:       file.ObjectID,
			OrganizationID: file.OrganizationID,
			FileName:       file.FileName,
			ContentType:    file.ContentType,
			SizeBytes:      file.SizeBytes,
			CreatedAt:      file.CreatedAt,
			SourceType:     file.SourceType,
			SourceID:       file.SourceID,
		})
	}
	return out
}
