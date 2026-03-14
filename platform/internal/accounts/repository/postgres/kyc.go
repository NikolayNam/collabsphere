package postgres

import (
	"context"
	"errors"
	"fmt"

	apperrors "github.com/NikolayNam/collabsphere/internal/accounts/application/errors"
	appports "github.com/NikolayNam/collabsphere/internal/accounts/application/ports"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *AccountRepo) GetAccountKYCProfile(ctx context.Context, accountID uuid.UUID) (*appports.AccountKYCProfileRecord, error) {
	var row appports.AccountKYCProfileRecord
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("kyc.account_profiles").
		Select("account_id, status, legal_name, country_code, document_number, residence_address, review_note, reviewer_account_id AS reviewer_account, submitted_at, reviewed_at, created_at, updated_at").
		Where("account_id = ?", accountID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *AccountRepo) UpsertAccountKYCProfile(ctx context.Context, accountID uuid.UUID, patch appports.AccountKYCProfilePatch) (*appports.AccountKYCProfileRecord, error) {
	payload := map[string]any{
		"account_id":          accountID,
		"status":              patch.Status,
		"legal_name":          patch.LegalName,
		"country_code":        patch.CountryCode,
		"document_number":     patch.DocumentNumber,
		"residence_address":   patch.ResidenceAddress,
		"review_note":         patch.ReviewNote,
		"reviewer_account_id": patch.ReviewerAccount,
		"submitted_at":        patch.SubmittedAt,
		"reviewed_at":         patch.ReviewedAt,
		"updated_at":          patch.UpdatedAt,
		"created_at":          patch.UpdatedAt,
	}
	db := r.dbFrom(ctx).WithContext(ctx)
	if err := db.Table("kyc.account_profiles").
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "account_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"status":              patch.Status,
				"legal_name":          patch.LegalName,
				"country_code":        patch.CountryCode,
				"document_number":     patch.DocumentNumber,
				"residence_address":   patch.ResidenceAddress,
				"review_note":         patch.ReviewNote,
				"reviewer_account_id": patch.ReviewerAccount,
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
	return r.GetAccountKYCProfile(ctx, accountID)
}

func (r *AccountRepo) ListAccountKYCDocuments(ctx context.Context, accountID uuid.UUID) ([]appports.AccountKYCDocumentRecord, error) {
	var rows []appports.AccountKYCDocumentRecord
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("kyc.account_documents").
		Select("id, account_id, object_id, document_type, title, status, review_note, reviewer_account_id AS reviewer_account, created_at, updated_at, reviewed_at").
		Where("account_id = ? AND deleted_at IS NULL", accountID).
		Order("created_at DESC, id DESC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *AccountRepo) GetAccountKYCDocumentByObjectID(ctx context.Context, accountID, objectID uuid.UUID) (*appports.AccountKYCDocumentRecord, error) {
	var row appports.AccountKYCDocumentRecord
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("kyc.account_documents").
		Select("id, account_id, object_id, document_type, title, status, review_note, reviewer_account_id AS reviewer_account, created_at, updated_at, reviewed_at").
		Where("account_id = ? AND object_id = ? AND deleted_at IS NULL", accountID, objectID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *AccountRepo) CreateAccountKYCDocument(ctx context.Context, item appports.AccountKYCDocumentRecord) (*appports.AccountKYCDocumentRecord, error) {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	payload := map[string]any{
		"id":                  item.ID,
		"account_id":          item.AccountID,
		"object_id":           item.ObjectID,
		"document_type":       item.DocumentType,
		"title":               item.Title,
		"status":              item.Status,
		"review_note":         item.ReviewNote,
		"reviewer_account_id": item.ReviewerAccount,
		"created_at":          item.CreatedAt,
	}
	if err := r.dbFrom(ctx).WithContext(ctx).Table("kyc.account_documents").Create(payload).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
		}
		if isForeignKeyViolation(err) {
			return nil, apperrors.InvalidInput("KYC document object does not exist")
		}
		return nil, err
	}
	return r.GetAccountKYCDocumentByObjectID(ctx, item.AccountID, item.ObjectID)
}
