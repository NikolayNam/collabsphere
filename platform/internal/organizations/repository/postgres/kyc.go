package postgres

import (
	"context"
	"errors"
	"fmt"

	apperrors "github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	appports "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *OrganizationRepo) GetOrganizationKYCProfile(ctx context.Context, organizationID uuid.UUID) (*appports.OrganizationKYCProfileRecord, error) {
	var row appports.OrganizationKYCProfileRecord
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("kyc.organization_profiles").
		Select("organization_id, status, legal_name, country_code, registration_number, tax_id, review_note, reviewer_account_id, submitted_at, reviewed_at, created_at, updated_at").
		Where("organization_id = ?", organizationID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *OrganizationRepo) UpsertOrganizationKYCProfile(ctx context.Context, organizationID uuid.UUID, patch appports.OrganizationKYCProfilePatch) (*appports.OrganizationKYCProfileRecord, error) {
	payload := map[string]any{
		"organization_id":     organizationID,
		"status":              patch.Status,
		"legal_name":          patch.LegalName,
		"country_code":        patch.CountryCode,
		"registration_number": patch.RegistrationNumber,
		"tax_id":              patch.TaxID,
		"review_note":         patch.ReviewNote,
		"reviewer_account_id": patch.ReviewerAccountID,
		"submitted_at":        patch.SubmittedAt,
		"reviewed_at":         patch.ReviewedAt,
		"updated_at":          patch.UpdatedAt,
		"created_at":          patch.UpdatedAt,
	}
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("kyc.organization_profiles").
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "organization_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"status":              patch.Status,
				"legal_name":          patch.LegalName,
				"country_code":        patch.CountryCode,
				"registration_number": patch.RegistrationNumber,
				"tax_id":              patch.TaxID,
				"review_note":         patch.ReviewNote,
				"reviewer_account_id": patch.ReviewerAccountID,
				"submitted_at":        patch.SubmittedAt,
				"reviewed_at":         patch.ReviewedAt,
				"updated_at":          patch.UpdatedAt,
			}),
		}).
		Create(payload).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
		}
		return nil, err
	}
	return r.GetOrganizationKYCProfile(ctx, organizationID)
}

func (r *OrganizationRepo) ListOrganizationKYCDocuments(ctx context.Context, organizationID uuid.UUID) ([]appports.OrganizationKYCDocumentRecord, error) {
	var rows []appports.OrganizationKYCDocumentRecord
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("kyc.organization_documents").
		Select("id, organization_id, object_id, document_type, title, status, review_note, reviewer_account_id, created_at, updated_at, reviewed_at").
		Where("organization_id = ? AND deleted_at IS NULL", organizationID).
		Order("created_at DESC, id DESC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *OrganizationRepo) GetOrganizationKYCDocumentByObjectID(ctx context.Context, organizationID, objectID uuid.UUID) (*appports.OrganizationKYCDocumentRecord, error) {
	var row appports.OrganizationKYCDocumentRecord
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("kyc.organization_documents").
		Select("id, organization_id, object_id, document_type, title, status, review_note, reviewer_account_id, created_at, updated_at, reviewed_at").
		Where("organization_id = ? AND object_id = ? AND deleted_at IS NULL", organizationID, objectID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *OrganizationRepo) CreateOrganizationKYCDocument(ctx context.Context, item appports.OrganizationKYCDocumentRecord) (*appports.OrganizationKYCDocumentRecord, error) {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	payload := map[string]any{
		"id":                  item.ID,
		"organization_id":     item.OrganizationID,
		"object_id":           item.ObjectID,
		"document_type":       item.DocumentType,
		"title":               item.Title,
		"status":              item.Status,
		"review_note":         item.ReviewNote,
		"reviewer_account_id": item.ReviewerAccountID,
		"created_at":          item.CreatedAt,
	}
	if err := r.dbFrom(ctx).WithContext(ctx).Table("kyc.organization_documents").Create(payload).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
		}
		if isForeignKeyViolation(err) {
			return nil, apperrors.InvalidInput("KYC document object does not exist")
		}
		return nil, err
	}
	return r.GetOrganizationKYCDocumentByObjectID(ctx, item.OrganizationID, item.ObjectID)
}
