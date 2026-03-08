package postgres

import (
	"context"
	"errors"
	"fmt"

	apperrors "github.com/NikolayNam/collabsphere/internal/accounts/application/errors"
	appports "github.com/NikolayNam/collabsphere/internal/accounts/application/ports"
	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres/mapper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *AccountRepo) Create(ctx context.Context, account *domain.Account) error {
	if account == nil {
		return errors.New("account is nil")
	}

	accountModel := mapper.ToDBAccountForCreate(account)
	credentialModel := mapper.ToDBPasswordCredentialForCreate(account)
	if accountModel == nil || credentialModel == nil {
		return errors.New("db account model is nil")
	}

	db := r.dbFrom(ctx).WithContext(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(accountModel).Error; err != nil {
			if isUniqueViolation(err) {
				return fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
			}
			return err
		}

		if err := tx.Create(credentialModel).Error; err != nil {
			if isUniqueViolation(err) {
				return fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
			}
			return err
		}

		return nil
	})
}

func (r *AccountRepo) UpdateProfile(ctx context.Context, id domain.AccountID, patch domain.AccountProfilePatch) (*domain.Account, error) {
	if id.IsZero() {
		return nil, apperrors.InvalidInput("Account ID is required")
	}

	current, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}
	if err := current.ApplyProfilePatch(patch); err != nil {
		return nil, fmt.Errorf("%w: %w", apperrors.ErrValidation, err)
	}

	db := r.dbFrom(ctx).WithContext(ctx)
	if patch.AvatarObjectID != nil && !patch.ClearAvatar {
		valid, err := r.accountAvatarExists(db, *patch.AvatarObjectID)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, apperrors.InvalidInput("Avatar object does not exist or is not a personal object")
		}
	}

	updates := map[string]any{
		"display_name":     current.DisplayName(),
		"bio":              current.Bio(),
		"phone":            current.Phone(),
		"locale":           current.Locale(),
		"timezone":         current.Timezone(),
		"website":          current.Website(),
		"updated_at":       current.UpdatedAt(),
		"avatar_object_id": current.AvatarObjectID(),
	}

	result := db.Table("iam.accounts").Where("id = ?", id.UUID()).Updates(updates)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return nil, fmt.Errorf("%w: %w", apperrors.ErrConflict, result.Error)
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return r.GetByID(ctx, id)
}

func (r *AccountRepo) CreateStorageObject(ctx context.Context, object appports.StorageObject) error {
	return r.dbFrom(ctx).WithContext(ctx).Table("storage.objects").Create(map[string]any{
		"id":              object.ID,
		"organization_id": object.OrganizationID,
		"bucket":          object.Bucket,
		"object_key":      object.ObjectKey,
		"file_name":       object.FileName,
		"content_type":    object.ContentType,
		"size_bytes":      object.SizeBytes,
		"checksum_sha256": object.ChecksumSHA256,
		"created_at":      object.CreatedAt,
	}).Error
}

func (r *AccountRepo) accountAvatarExists(db *gorm.DB, objectID uuid.UUID) (bool, error) {
	var n int64
	err := db.Table("storage.objects").
		Where("id = ? AND organization_id IS NULL AND deleted_at IS NULL", objectID).
		Limit(1).
		Count(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
