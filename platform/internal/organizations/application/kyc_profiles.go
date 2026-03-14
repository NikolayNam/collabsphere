package application

import (
	"context"
	"strings"
	"time"

	apperrors "github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	uploaddomain "github.com/NikolayNam/collabsphere/internal/uploads/domain"
	sharedkyc "github.com/NikolayNam/collabsphere/shared/kyc"
	"github.com/google/uuid"
)

type GetOrganizationKYCProfileQuery struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
}

type UpdateOrganizationKYCProfileCmd struct {
	OrganizationID     domain.OrganizationID
	ActorAccountID     uuid.UUID
	Status             *string
	LegalName          *string
	CountryCode        *string
	RegistrationNumber *string
	TaxID              *string
}

type CreateOrganizationKYCDocumentUploadCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	DocumentType   string
	Title          *string
	FileName       string
	ContentType    *string
	SizeBytes      *int64
	ChecksumSHA256 *string
}

type CreateOrganizationKYCDocumentUploadResult struct {
	UploadID     uuid.UUID
	ObjectID     uuid.UUID
	Bucket       string
	ObjectKey    string
	UploadURL    string
	ExpiresAt    time.Time
	FileName     string
	SizeBytes    int64
	DocumentType string
	Title        string
}

type CompleteOrganizationKYCDocumentUploadCmd struct {
	OrganizationID domain.OrganizationID
	ActorAccountID uuid.UUID
	UploadID       uuid.UUID
}

func (s *Service) GetOrganizationKYCProfile(ctx context.Context, q GetOrganizationKYCProfileQuery) (*ports.OrganizationKYCProfileRecord, []ports.OrganizationKYCDocumentRecord, error) {
	if err := s.requireOrganizationAccess(ctx, q.OrganizationID, q.ActorAccountID, true); err != nil {
		return nil, nil, err
	}
	profile, err := s.repo.GetOrganizationKYCProfile(ctx, q.OrganizationID.UUID())
	if err != nil {
		return nil, nil, err
	}
	docs, err := s.repo.ListOrganizationKYCDocuments(ctx, q.OrganizationID.UUID())
	if err != nil {
		return nil, nil, err
	}
	if profile == nil {
		now := s.clock.Now()
		profile = &ports.OrganizationKYCProfileRecord{
			OrganizationID: q.OrganizationID.UUID(),
			Status:         string(sharedkyc.StatusDraft),
			CreatedAt:      now,
			UpdatedAt:      now,
		}
	}
	return profile, docs, nil
}

func (s *Service) UpdateOrganizationKYCProfile(ctx context.Context, cmd UpdateOrganizationKYCProfileCmd) (*ports.OrganizationKYCProfileRecord, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	current, err := s.repo.GetOrganizationKYCProfile(ctx, cmd.OrganizationID.UUID())
	if err != nil {
		return nil, err
	}
	nextStatus := sharedkyc.StatusDraft
	if current != nil {
		parsed, ok := sharedkyc.ParseStatus(current.Status)
		if ok {
			nextStatus = parsed
		}
	}
	var submittedAt *time.Time
	if current != nil {
		submittedAt = current.SubmittedAt
	}
	if cmd.Status != nil {
		parsed, ok := sharedkyc.ParseStatus(*cmd.Status)
		if !ok {
			return nil, apperrors.InvalidInput("KYC status is invalid")
		}
		nextStatus = parsed
		now := s.clock.Now()
		switch parsed {
		case sharedkyc.StatusSubmitted:
			submittedAt = &now
		case sharedkyc.StatusDraft:
			submittedAt = nil
		}
	}
	now := s.clock.Now()
	return s.repo.UpsertOrganizationKYCProfile(ctx, cmd.OrganizationID.UUID(), ports.OrganizationKYCProfilePatch{
		Status:             string(nextStatus),
		LegalName:          cmd.LegalName,
		CountryCode:        cmd.CountryCode,
		RegistrationNumber: cmd.RegistrationNumber,
		TaxID:              cmd.TaxID,
		ReviewNote:         nil,
		ReviewerAccountID:  nil,
		SubmittedAt:        submittedAt,
		ReviewedAt:         nil,
		UpdatedAt:          now,
	})
}

func (s *Service) CreateOrganizationKYCDocumentUpload(ctx context.Context, cmd CreateOrganizationKYCDocumentUploadCmd) (*CreateOrganizationKYCDocumentUploadResult, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if s.storage == nil || s.bucket == "" || s.uploads == nil {
		return nil, fault.Unavailable("KYC document upload is unavailable")
	}
	fileName := sanitizeFileName(strings.TrimSpace(cmd.FileName), "kyc-document.bin")
	documentType := strings.TrimSpace(cmd.DocumentType)
	if documentType == "" {
		return nil, apperrors.InvalidInput("documentType is required")
	}
	title := fileName
	if cmd.Title != nil && strings.TrimSpace(*cmd.Title) != "" {
		title = strings.TrimSpace(*cmd.Title)
	}
	sizeBytes := int64(0)
	if cmd.SizeBytes != nil {
		if *cmd.SizeBytes < 0 {
			return nil, apperrors.InvalidInput("sizeBytes must be non-negative")
		}
		sizeBytes = *cmd.SizeBytes
	}
	now := s.clock.Now()
	objectID := uuid.New()
	uploadID := uuid.New()
	objectKey := strings.Join([]string{
		"organizations",
		"kyc",
		cmd.OrganizationID.UUID().String(),
		now.UTC().Format("2006"),
		now.UTC().Format("01"),
		now.UTC().Format("02"),
		objectID.String(),
		fileName,
	}, "/")
	object := ports.StorageObject{
		ID:             objectID,
		OrganizationID: ptrUUID(cmd.OrganizationID.UUID()),
		Bucket:         s.bucket,
		ObjectKey:      objectKey,
		FileName:       fileName,
		ContentType:    normalizeOptionalString(cmd.ContentType),
		SizeBytes:      sizeBytes,
		ChecksumSHA256: normalizeOptionalString(cmd.ChecksumSHA256),
		CreatedAt:      now,
	}
	if err := s.repo.CreateStorageObject(ctx, object); err != nil {
		return nil, fault.Internal("Create KYC object failed", fault.WithCause(err))
	}
	uploadURL, expiresAt, err := s.storage.PresignPutObject(ctx, object.Bucket, object.ObjectKey)
	if err != nil {
		return nil, fault.Internal("Presign KYC upload failed", fault.WithCause(err))
	}
	if err := s.uploads.Create(ctx, &uploaddomain.Upload{
		ID:                 uploadID,
		OrganizationID:     ptrUUID(cmd.OrganizationID.UUID()),
		ObjectID:           objectID,
		CreatedByAccountID: cmd.ActorAccountID,
		Purpose:            uploaddomain.PurposeOrganizationKYCDocument,
		Status:             uploaddomain.StatusPending,
		Bucket:             object.Bucket,
		ObjectKey:          object.ObjectKey,
		FileName:           object.FileName,
		ContentType:        object.ContentType,
		DeclaredSizeBytes:  object.SizeBytes,
		ChecksumSHA256:     object.ChecksumSHA256,
		Metadata: map[string]any{
			"documentType": documentType,
			"title":        title,
		},
		ExpiresAt: &expiresAt,
		CreatedAt: now,
	}); err != nil {
		return nil, fault.Internal("Create KYC upload failed", fault.WithCause(err))
	}
	return &CreateOrganizationKYCDocumentUploadResult{
		UploadID:     uploadID,
		ObjectID:     objectID,
		Bucket:       object.Bucket,
		ObjectKey:    object.ObjectKey,
		UploadURL:    uploadURL,
		ExpiresAt:    expiresAt,
		FileName:     fileName,
		SizeBytes:    sizeBytes,
		DocumentType: documentType,
		Title:        title,
	}, nil
}

func (s *Service) CompleteOrganizationKYCDocumentUpload(ctx context.Context, cmd CompleteOrganizationKYCDocumentUploadCmd) (*ports.OrganizationKYCDocumentRecord, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if cmd.UploadID == uuid.Nil {
		return nil, apperrors.InvalidInput("uploadId is required")
	}
	if s.uploads == nil || s.storage == nil {
		return nil, fault.Unavailable("KYC document upload is unavailable")
	}
	upload, err := s.uploads.GetByID(ctx, cmd.UploadID)
	if err != nil {
		return nil, fault.Internal("Load KYC upload failed", fault.WithCause(err))
	}
	if upload == nil || upload.Purpose != uploaddomain.PurposeOrganizationKYCDocument || upload.OrganizationID == nil || *upload.OrganizationID != cmd.OrganizationID.UUID() {
		return nil, fault.NotFound("KYC upload not found")
	}
	if upload.Status == uploaddomain.StatusReady {
		item, err := s.repo.GetOrganizationKYCDocumentByObjectID(ctx, cmd.OrganizationID.UUID(), upload.ObjectID)
		if err != nil {
			return nil, err
		}
		if item == nil {
			return nil, fault.NotFound("KYC document not found")
		}
		return item, nil
	}
	reader, err := s.storage.ReadObject(ctx, upload.Bucket, upload.ObjectKey)
	if err != nil {
		_, _ = s.uploads.MarkFailed(ctx, cmd.UploadID, "storage_object_missing", "Uploaded file is not available in storage", s.clock.Now())
		return nil, fault.Validation("Uploaded file is not available")
	}
	_ = reader.Close()
	documentType := strings.TrimSpace(kycMetadataString(upload.Metadata, "documentType"))
	if documentType == "" {
		return nil, apperrors.InvalidInput("documentType is required")
	}
	title := strings.TrimSpace(kycMetadataString(upload.Metadata, "title"))
	if title == "" {
		title = upload.FileName
	}
	now := s.clock.Now()
	record, err := s.repo.CreateOrganizationKYCDocument(ctx, ports.OrganizationKYCDocumentRecord{
		OrganizationID: cmd.OrganizationID.UUID(),
		ObjectID:       upload.ObjectID,
		DocumentType:   documentType,
		Title:          title,
		Status:         string(sharedkyc.DocumentStatusUploaded),
		CreatedAt:      now,
	})
	if err != nil {
		return nil, err
	}
	actualSize := int64Ptr(upload.DeclaredSizeBytes)
	_, err = s.uploads.MarkReady(ctx, upload.ID, actualSize, uploaddomain.ResultKindOrganizationKYCDocument, record.ID, now, now)
	if err != nil {
		return nil, fault.Internal("Finalize KYC upload failed", fault.WithCause(err))
	}
	return record, nil
}

func ptrUUID(id uuid.UUID) *uuid.UUID {
	return &id
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func kycMetadataString(metadata map[string]any, key string) string {
	if len(metadata) == 0 {
		return ""
	}
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}
