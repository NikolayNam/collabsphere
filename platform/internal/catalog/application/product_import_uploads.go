package application

import (
	"context"

	catalogaccess "github.com/NikolayNam/collabsphere/internal/catalog/application/access"
	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/run_product_import"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	uploaddomain "github.com/NikolayNam/collabsphere/internal/uploads/domain"
	"github.com/google/uuid"
)

func (s *Service) CompleteProductImportUpload(ctx context.Context, cmd CompleteProductImportUploadCmd) (*ProductImportView, error) {
	if cmd.UploadID == uuid.Nil {
		return nil, catalogerrors.ProductImportFileInvalid("uploadId is required")
	}
	if err := catalogaccess.RequireOrganizationAccess(ctx, s.organizations, s.memberships, s.roleResolver, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if s.uploads == nil || s.storage == nil || s.runImport == nil {
		return nil, catalogerrors.ProductImportUnavailable()
	}

	upload, err := s.uploads.GetByID(ctx, cmd.UploadID)
	if err != nil {
		return nil, fault.Internal("Load product import upload failed", fault.WithCause(err))
	}
	if upload == nil || upload.Purpose != uploaddomain.PurposeProductImport || upload.OrganizationID == nil || *upload.OrganizationID != cmd.OrganizationID.UUID() {
		return nil, fault.NotFound("Product import upload not found")
	}
	if upload.Status == uploaddomain.StatusReady {
		return s.resolveCompletedProductImportUpload(ctx, cmd, upload)
	}
	if upload.Status == uploaddomain.StatusFailed {
		return nil, fault.Conflict("Product import upload is already in failed state")
	}

	reader, err := s.storage.ReadObject(ctx, upload.Bucket, upload.ObjectKey)
	if err != nil {
		s.markProductImportUploadFailed(ctx, cmd.UploadID, "storage_object_missing", "Uploaded file is not available in storage")
		return nil, catalogerrors.ProductImportFileInvalid("Uploaded file is not available")
	}
	_ = reader.Close()

	existingBatch, err := s.repo.GetProductImportBatchBySourceObjectID(ctx, cmd.OrganizationID, upload.ObjectID)
	if err != nil {
		return nil, err
	}
	if existingBatch != nil {
		completedAt := s.clock.Now()
		_, err = s.uploads.MarkReady(ctx, upload.ID, uploadInt64Ptr(upload.DeclaredSizeBytes), uploaddomain.ResultKindProductImportBatch, existingBatch.ID, completedAt, completedAt)
		if err != nil {
			return nil, err
		}
		return s.getImport.Handle(ctx, GetProductImportQuery{
			OrganizationID: cmd.OrganizationID,
			ActorAccountID: cmd.ActorAccountID,
			BatchID:        existingBatch.ID,
		})
	}

	view, err := s.runImport.Handle(ctx, run_product_import.Command{
		OrganizationID: cmd.OrganizationID,
		ActorAccountID: cmd.ActorAccountID,
		SourceObjectID: upload.ObjectID,
		Mode:           cmd.Mode,
	})
	if err != nil {
		return nil, err
	}
	completedAt := s.clock.Now()
	_, err = s.uploads.MarkReady(ctx, upload.ID, uploadInt64Ptr(upload.DeclaredSizeBytes), uploaddomain.ResultKindProductImportBatch, view.Batch.ID, completedAt, completedAt)
	if err != nil {
		return nil, err
	}
	return view, nil
}

func (s *Service) resolveCompletedProductImportUpload(ctx context.Context, cmd CompleteProductImportUploadCmd, upload *uploaddomain.Upload) (*ProductImportView, error) {
	if upload == nil {
		return nil, fault.NotFound("Product import upload not found")
	}
	if upload.ResultID != nil && upload.ResultKind != nil && *upload.ResultKind == uploaddomain.ResultKindProductImportBatch {
		return s.getImport.Handle(ctx, GetProductImportQuery{
			OrganizationID: cmd.OrganizationID,
			ActorAccountID: cmd.ActorAccountID,
			BatchID:        *upload.ResultID,
		})
	}
	existingBatch, err := s.repo.GetProductImportBatchBySourceObjectID(ctx, cmd.OrganizationID, upload.ObjectID)
	if err != nil {
		return nil, err
	}
	if existingBatch == nil {
		return nil, fault.NotFound("Product import batch not found")
	}
	return s.getImport.Handle(ctx, GetProductImportQuery{
		OrganizationID: cmd.OrganizationID,
		ActorAccountID: cmd.ActorAccountID,
		BatchID:        existingBatch.ID,
	})
}

func (s *Service) markProductImportUploadFailed(ctx context.Context, uploadID uuid.UUID, code, message string) {
	if s.uploads == nil || uploadID == uuid.Nil {
		return
	}
	_, _ = s.uploads.MarkFailed(ctx, uploadID, code, message, s.clock.Now())
}

func uploadInt64Ptr(value int64) *int64 {
	return &value
}

func cloneOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := *value
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
