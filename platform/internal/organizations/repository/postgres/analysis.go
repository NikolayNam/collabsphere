package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	apperrors "github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	appports "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/mapper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type legalDocumentAnalysisLeaseRow struct {
	JobID          uuid.UUID `gorm:"column:job_id"`
	DocumentID     uuid.UUID `gorm:"column:document_id"`
	OrganizationID uuid.UUID `gorm:"column:organization_id"`
	ObjectID       uuid.UUID `gorm:"column:object_id"`
	Bucket         string    `gorm:"column:bucket"`
	ObjectKey      string    `gorm:"column:object_key"`
	FileName       string    `gorm:"column:file_name"`
	MimeType       *string   `gorm:"column:mime_type"`
	Provider       string    `gorm:"column:provider"`
	Attempts       int       `gorm:"column:attempts"`
}

func (r *OrganizationRepo) GetOrganizationLegalDocumentByID(ctx context.Context, organizationID domain.OrganizationID, documentID uuid.UUID) (*domain.OrganizationLegalDocument, error) {
	if organizationID.IsZero() {
		return nil, apperrors.InvalidInput("Organization ID is required")
	}
	if documentID == uuid.Nil {
		return nil, apperrors.InvalidInput("Document ID is required")
	}
	var model dbmodel.OrganizationLegalDocument
	err := r.dbFrom(ctx).WithContext(ctx).
		Where("organization_id = ? AND id = ? AND deleted_at IS NULL", organizationID.UUID(), documentID).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToDomainOrganizationLegalDocument(&model)
}

func (r *OrganizationRepo) GetOrganizationLegalDocumentByObjectID(ctx context.Context, organizationID domain.OrganizationID, objectID uuid.UUID) (*domain.OrganizationLegalDocument, error) {
	if organizationID.IsZero() {
		return nil, apperrors.InvalidInput("Organization ID is required")
	}
	if objectID == uuid.Nil {
		return nil, apperrors.InvalidInput("Object ID is required")
	}
	var model dbmodel.OrganizationLegalDocument
	err := r.dbFrom(ctx).WithContext(ctx).
		Where("organization_id = ? AND object_id = ? AND deleted_at IS NULL", organizationID.UUID(), objectID).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToDomainOrganizationLegalDocument(&model)
}

func (r *OrganizationRepo) GetOrganizationLegalDocumentAnalysis(ctx context.Context, organizationID domain.OrganizationID, documentID uuid.UUID) (*domain.OrganizationLegalDocumentAnalysis, error) {
	if organizationID.IsZero() {
		return nil, apperrors.InvalidInput("Organization ID is required")
	}
	if documentID == uuid.Nil {
		return nil, apperrors.InvalidInput("Document ID is required")
	}
	var model dbmodel.OrganizationLegalDocumentAnalysis
	err := r.dbFrom(ctx).WithContext(ctx).
		Where("organization_id = ? AND document_id = ?", organizationID.UUID(), documentID).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToDomainOrganizationLegalDocumentAnalysis(&model)
}

func (r *OrganizationRepo) EnsureOrganizationLegalDocumentAnalysis(ctx context.Context, document *domain.OrganizationLegalDocument, provider string, now time.Time) error {
	if document == nil {
		return apperrors.InvalidInput("Organization legal document is required")
	}
	if strings.TrimSpace(provider) == "" {
		provider = "generic-http"
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return r.dbFrom(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			INSERT INTO org.organization_legal_document_analysis (
				id, document_id, organization_id, status, provider, extracted_fields_json, requested_at, updated_at
			)
			VALUES (@id, @document_id, @organization_id, 'pending', @provider, '{}'::jsonb, @requested_at, @updated_at)
			ON CONFLICT (document_id)
			DO UPDATE SET status = 'pending',
			              provider = EXCLUDED.provider,
			              requested_at = EXCLUDED.requested_at,
			              started_at = NULL,
			              completed_at = NULL,
			              updated_at = EXCLUDED.updated_at,
			              last_error = NULL
		`, map[string]any{
			"id":              uuid.New(),
			"document_id":     document.ID(),
			"organization_id": document.OrganizationID().UUID(),
			"provider":        provider,
			"requested_at":    now,
			"updated_at":      now,
		}).Error; err != nil {
			return err
		}
		return tx.Exec(`
			INSERT INTO integration.organization_document_analysis_jobs (
				id, document_id, status, provider, attempts, available_at, created_at, updated_at
			)
			VALUES (@id, @document_id, 'pending', @provider, 0, @available_at, @created_at, @updated_at)
			ON CONFLICT (document_id)
			DO UPDATE SET status = 'pending',
			              provider = EXCLUDED.provider,
			              available_at = EXCLUDED.available_at,
			              leased_until = NULL,
			              completed_at = NULL,
			              last_error = NULL,
			              updated_at = EXCLUDED.updated_at
		`, map[string]any{
			"id":           uuid.New(),
			"document_id":  document.ID(),
			"provider":     provider,
			"available_at": now,
			"created_at":   now,
			"updated_at":   now,
		}).Error
	})
}

func (r *OrganizationRepo) LeaseNextOrganizationLegalDocumentAnalysisJob(ctx context.Context, now time.Time, leaseFor time.Duration) (*appports.LegalDocumentAnalysisLease, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	leasedUntil := now.Add(leaseFor)
	var leased legalDocumentAnalysisLeaseRow
	err := r.dbFrom(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row legalDocumentAnalysisLeaseRow
		if err := tx.Table("integration.organization_document_analysis_jobs AS j").
			Select("j.id AS job_id, j.document_id, ld.organization_id, ld.object_id, so.bucket, so.object_key, so.file_name, so.content_type AS mime_type, j.provider, j.attempts").
			Joins("JOIN org.organization_legal_documents AS ld ON ld.id = j.document_id").
			Joins("JOIN storage.objects AS so ON so.id = ld.object_id").
			Where("ld.deleted_at IS NULL AND j.status IN ('pending', 'failed') AND j.available_at <= ? AND (j.leased_until IS NULL OR j.leased_until < ?)", now, now).
			Order("j.available_at ASC, j.created_at ASC").
			Take(&row).Error; err != nil {
			return err
		}
		if err := tx.Table("integration.organization_document_analysis_jobs").Where("id = ?", row.JobID).Updates(map[string]any{
			"status":       "leased",
			"leased_until": leasedUntil,
			"attempts":     row.Attempts + 1,
			"updated_at":   now,
		}).Error; err != nil {
			return err
		}
		if err := tx.Table("org.organization_legal_document_analysis").Where("document_id = ?", row.DocumentID).Updates(map[string]any{
			"status":     string(domain.OrganizationLegalDocumentAnalysisStatusProcessing),
			"started_at": now,
			"updated_at": now,
			"last_error": nil,
		}).Error; err != nil {
			return err
		}
		leased = row
		leased.Attempts = row.Attempts + 1
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &appports.LegalDocumentAnalysisLease{
		JobID:          leased.JobID,
		DocumentID:     leased.DocumentID,
		OrganizationID: leased.OrganizationID,
		ObjectID:       leased.ObjectID,
		Bucket:         leased.Bucket,
		ObjectKey:      leased.ObjectKey,
		FileName:       leased.FileName,
		MimeType:       leased.MimeType,
		Provider:       leased.Provider,
		Attempts:       leased.Attempts,
	}, nil
}

func (r *OrganizationRepo) CompleteOrganizationLegalDocumentAnalysisJob(ctx context.Context, jobID, documentID uuid.UUID, provider string, result appports.LegalDocumentAnalysisResult, completedAt time.Time) error {
	if completedAt.IsZero() {
		completedAt = time.Now().UTC()
	}
	if len(result.ExtractedFieldsJSON) == 0 {
		result.ExtractedFieldsJSON = []byte(`{}`)
	}
	return r.dbFrom(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			UPDATE org.organization_legal_document_analysis
			SET status = @status,
			    provider = @provider,
			    extracted_text = @extracted_text,
			    summary = @summary,
			    extracted_fields_json = CAST(@fields_json AS jsonb),
			    detected_document_type = @detected_document_type,
			    confidence_score = @confidence_score,
			    completed_at = @completed_at,
			    updated_at = @updated_at,
			    last_error = NULL
			WHERE document_id = @document_id
		`, map[string]any{
			"status":                 string(domain.OrganizationLegalDocumentAnalysisStatusCompleted),
			"provider":               provider,
			"extracted_text":         strings.TrimSpace(result.ExtractedText),
			"summary":                result.Summary,
			"fields_json":            string(result.ExtractedFieldsJSON),
			"detected_document_type": result.DetectedDocumentType,
			"confidence_score":       result.ConfidenceScore,
			"completed_at":           completedAt,
			"updated_at":             completedAt,
			"document_id":            documentID,
		}).Error; err != nil {
			return err
		}
		return tx.Table("integration.organization_document_analysis_jobs").Where("id = ?", jobID).Updates(map[string]any{
			"status":       "completed",
			"completed_at": completedAt,
			"leased_until": nil,
			"updated_at":   completedAt,
		}).Error
	})
}

func (r *OrganizationRepo) FailOrganizationLegalDocumentAnalysisJob(ctx context.Context, jobID, documentID uuid.UUID, provider, errMessage string, retryAt time.Time) error {
	if retryAt.IsZero() {
		retryAt = time.Now().UTC().Add(30 * time.Second)
	}
	errMessage = strings.TrimSpace(errMessage)
	if errMessage == "" {
		errMessage = "document analysis failed"
	}
	now := time.Now().UTC()
	return r.dbFrom(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("org.organization_legal_document_analysis").Where("document_id = ?", documentID).Updates(map[string]any{
			"status":     string(domain.OrganizationLegalDocumentAnalysisStatusFailed),
			"provider":   provider,
			"updated_at": now,
			"last_error": errMessage,
		}).Error; err != nil {
			return err
		}
		return tx.Table("integration.organization_document_analysis_jobs").Where("id = ?", jobID).Updates(map[string]any{
			"status":       "failed",
			"available_at": retryAt,
			"leased_until": nil,
			"last_error":   errMessage,
			"updated_at":   now,
		}).Error
	})
}
