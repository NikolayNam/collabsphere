package application

import (
	"context"
	"strings"
	"time"

	apperrors "github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	uploaddomain "github.com/NikolayNam/collabsphere/internal/uploads/domain"
	"github.com/google/uuid"
)

func (s *Service) CompleteLegalDocumentUpload(ctx context.Context, cmd CompleteLegalDocumentUploadCmd) (*domain.OrganizationLegalDocument, error) {
	if err := s.requireOrganizationAccess(ctx, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if cmd.UploadID == uuid.Nil {
		return nil, apperrors.InvalidInput("uploadId is required")
	}
	if s.uploads == nil {
		return nil, fault.Unavailable("Upload tracking is unavailable")
	}
	if s.storage == nil {
		return nil, fault.Unavailable("File upload is unavailable")
	}

	upload, err := s.uploads.GetByID(ctx, cmd.UploadID)
	if err != nil {
		return nil, fault.Internal("Load legal document upload failed", fault.WithCause(err))
	}
	if upload == nil || upload.Purpose != uploaddomain.PurposeOrganizationLegalDocument || upload.OrganizationID == nil || *upload.OrganizationID != cmd.OrganizationID.UUID() {
		return nil, fault.NotFound("Legal document upload not found")
	}

	if upload.Status == uploaddomain.StatusReady {
		return s.resolveCompletedLegalDocumentUpload(ctx, cmd.OrganizationID, upload)
	}
	if upload.Status == uploaddomain.StatusFailed {
		return nil, fault.Conflict("Legal document upload is already in failed state")
	}

	reader, err := s.storage.ReadObject(ctx, upload.Bucket, upload.ObjectKey)
	if err != nil {
		s.markLegalDocumentUploadFailed(ctx, cmd.UploadID, "storage_object_missing", "Uploaded file is not available in storage")
		return nil, fault.Validation("Uploaded file is not available")
	}
	_ = reader.Close()

	documentType := metadataString(upload.Metadata, "documentType")
	if documentType == "" {
		s.markLegalDocumentUploadFailed(ctx, cmd.UploadID, "missing_document_type", "Upload metadata does not contain documentType")
		return nil, apperrors.InvalidInput("documentType is required")
	}
	title := normalizeLegalDocumentTitle(metadataString(upload.Metadata, "title"), upload.FileName)
	actualSize := int64Ptr(upload.DeclaredSizeBytes)
	completedAt := s.clock.Now()

	var document *domain.OrganizationLegalDocument
	err = s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		existing, err := s.repo.GetOrganizationLegalDocumentByObjectID(ctx, cmd.OrganizationID, upload.ObjectID)
		if err != nil {
			return err
		}
		if existing != nil {
			document = existing
		} else {
			document, err = s.AddOrganizationLegalDocument(ctx, AddOrganizationLegalDocumentCmd{
				OrganizationID: cmd.OrganizationID,
				ActorAccountID: cmd.ActorAccountID,
				DocumentType:   documentType,
				ObjectID:       upload.ObjectID,
				Title:          title,
			})
			if err != nil {
				return err
			}
		}
		_, err = s.uploads.MarkReady(ctx, upload.ID, actualSize, uploaddomain.ResultKindOrganizationLegalDocument, document.ID(), completedAt, completedAt)
		return err
	})
	if err != nil {
		return nil, err
	}
	return document, nil
}

func (s *Service) resolveCompletedLegalDocumentUpload(ctx context.Context, organizationID domain.OrganizationID, upload *uploaddomain.Upload) (*domain.OrganizationLegalDocument, error) {
	if upload == nil {
		return nil, fault.NotFound("Legal document upload not found")
	}
	if upload.ResultID != nil && upload.ResultKind != nil && *upload.ResultKind == uploaddomain.ResultKindOrganizationLegalDocument {
		document, err := s.repo.GetOrganizationLegalDocumentByID(ctx, organizationID, *upload.ResultID)
		if err != nil {
			return nil, err
		}
		if document != nil {
			return document, nil
		}
	}
	document, err := s.repo.GetOrganizationLegalDocumentByObjectID(ctx, organizationID, upload.ObjectID)
	if err != nil {
		return nil, err
	}
	if document == nil {
		return nil, fault.NotFound("Organization legal document not found")
	}
	return document, nil
}

func (s *Service) markLegalDocumentUploadFailed(ctx context.Context, uploadID uuid.UUID, code, message string) {
	if s.uploads == nil || uploadID == uuid.Nil {
		return
	}
	_, _ = s.uploads.MarkFailed(ctx, uploadID, code, message, s.clock.Now())
}

func normalizeLegalDocumentTitle(title, fileName string) string {
	trimmed := strings.TrimSpace(title)
	if trimmed != "" {
		return trimmed
	}
	trimmed = strings.TrimSpace(fileName)
	if trimmed == "" {
		return "Document"
	}
	return trimmed
}

func metadataString(metadata map[string]any, key string) string {
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

func int64Ptr(value int64) *int64 {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
