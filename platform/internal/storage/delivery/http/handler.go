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

func (h *Handler) DownloadMyAvatar(ctx context.Context, _ *dto.DownloadMyAvatarInput) (*dto.DownloadObjectResponse, error) {
	result, err := h.svc.CreateMyAvatarDownload(ctx, storageapp.DownloadMyAvatarQuery{Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return downloadResponse(result), nil
}

func (h *Handler) DownloadOrganizationLogo(ctx context.Context, input *dto.DownloadOrganizationLogoInput) (*dto.DownloadObjectResponse, error) {
	organizationID, err := httpbind.ParseUUID(input.OrganizationID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	result, err := h.svc.CreateOrganizationLogoDownload(ctx, storageapp.DownloadOrganizationLogoQuery{OrganizationID: organizationID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return downloadResponse(result), nil
}

func (h *Handler) DownloadCooperationPriceList(ctx context.Context, input *dto.DownloadCooperationPriceListInput) (*dto.DownloadObjectResponse, error) {
	organizationID, err := httpbind.ParseUUID(input.OrganizationID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	result, err := h.svc.CreateCooperationPriceListDownload(ctx, storageapp.DownloadCooperationPriceListQuery{OrganizationID: organizationID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return downloadResponse(result), nil
}

func (h *Handler) DownloadOrganizationLegalDocument(ctx context.Context, input *dto.DownloadOrganizationLegalDocumentInput) (*dto.DownloadObjectResponse, error) {
	organizationID, err := httpbind.ParseUUID(input.OrganizationID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	documentID, err := httpbind.ParseUUID(input.DocumentID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	result, err := h.svc.CreateOrganizationLegalDocumentDownload(ctx, storageapp.DownloadOrganizationLegalDocumentQuery{OrganizationID: organizationID, DocumentID: documentID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return downloadResponse(result), nil
}

func (h *Handler) DownloadProductImportSource(ctx context.Context, input *dto.DownloadProductImportSourceInput) (*dto.DownloadObjectResponse, error) {
	organizationID, err := httpbind.ParseUUID(input.OrganizationID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	batchID, err := httpbind.ParseUUID(input.BatchID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	result, err := h.svc.CreateProductImportSourceDownload(ctx, storageapp.DownloadProductImportSourceQuery{OrganizationID: organizationID, BatchID: batchID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return downloadResponse(result), nil
}

func (h *Handler) DownloadChannelAttachment(ctx context.Context, input *dto.DownloadChannelAttachmentInput) (*dto.DownloadObjectResponse, error) {
	channelID, err := httpbind.ParseUUID(input.ChannelID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	objectID, err := httpbind.ParseUUID(input.ObjectID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	result, err := h.svc.CreateChannelAttachmentDownload(ctx, storageapp.DownloadChannelAttachmentQuery{ChannelID: channelID, ObjectID: objectID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return downloadResponse(result), nil
}

func (h *Handler) ListConferenceRecordings(ctx context.Context, input *dto.ListConferenceRecordingsInput) (*dto.ListConferenceRecordingsResponse, error) {
	conferenceID, err := httpbind.ParseUUID(input.ConferenceID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	items, err := h.svc.ListConferenceRecordings(ctx, storageapp.ListConferenceRecordingsQuery{ConferenceID: conferenceID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return listConferenceRecordingsResponse(items), nil
}

func (h *Handler) DownloadConferenceRecording(ctx context.Context, input *dto.DownloadConferenceRecordingInput) (*dto.DownloadObjectResponse, error) {
	conferenceID, err := httpbind.ParseUUID(input.ConferenceID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	recordingID, err := httpbind.ParseUUID(input.RecordingID, fault.Validation("Invalid identifier"))
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	result, err := h.svc.CreateConferenceRecordingDownload(ctx, storageapp.DownloadConferenceRecordingQuery{ConferenceID: conferenceID, RecordingID: recordingID, Actor: principal(ctx)})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return downloadResponse(result), nil
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

func downloadResponse(result *storageapp.DownloadObjectResult) *dto.DownloadObjectResponse {
	out := &dto.DownloadObjectResponse{Status: 200}
	out.Body.ObjectID = result.ObjectID
	out.Body.OrganizationID = result.OrganizationID
	out.Body.FileName = result.FileName
	out.Body.ContentType = result.ContentType
	out.Body.SizeBytes = result.SizeBytes
	out.Body.DownloadURL = result.DownloadURL
	out.Body.ExpiresAt = result.ExpiresAt
	out.Body.CreatedAt = result.CreatedAt
	return out
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

func listConferenceRecordingsResponse(items []storageapp.ConferenceRecordingFile) *dto.ListConferenceRecordingsResponse {
	out := &dto.ListConferenceRecordingsResponse{Status: 200}
	out.Body.Items = make([]dto.ConferenceRecordingItem, 0, len(items))
	for _, item := range items {
		out.Body.Items = append(out.Body.Items, dto.ConferenceRecordingItem{
			RecordingID:  item.RecordingID,
			ConferenceID: item.ConferenceID,
			ObjectID:     item.ObjectID,
			FileName:     item.FileName,
			ContentType:  item.ContentType,
			SizeBytes:    item.SizeBytes,
			CreatedAt:    item.CreatedAt,
			CreatedBy:    item.CreatedBy,
			DurationSec:  item.DurationSec,
			MimeType:     item.MimeType,
		})
	}
	return out
}
