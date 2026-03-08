package postgres

import (
	"context"
	"errors"
	"fmt"

	apperrors "github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	appports "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/mapper"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *OrganizationRepo) Create(ctx context.Context, organization *domain.Organization) error {
	if organization == nil {
		return apperrors.InvalidInput("Organization is required")
	}

	m := mapper.ToDBOrganizationForCreate(organization)
	if m == nil {
		return errors.New("db organization model is nil")
	}

	err := r.dbFrom(ctx).WithContext(ctx).Create(m).Error
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
		}
		return err
	}
	return nil
}

func (r *OrganizationRepo) UpdateProfile(ctx context.Context, id domain.OrganizationID, patch domain.OrganizationProfilePatch) (*domain.Organization, error) {
	if id.IsZero() {
		return nil, apperrors.InvalidInput("Organization ID is required")
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
	if patch.LogoObjectID != nil && !patch.ClearLogo {
		valid, err := r.organizationLogoExists(db, id.UUID(), *patch.LogoObjectID)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, apperrors.InvalidInput("Logo object does not exist or does not belong to organization")
		}
	}

	updates := map[string]any{
		"name":           current.Name(),
		"slug":           current.Slug(),
		"logo_object_id": current.LogoObjectID(),
		"description":    current.Description(),
		"website":        current.Website(),
		"primary_email":  current.PrimaryEmail(),
		"phone":          current.Phone(),
		"address":        current.Address(),
		"industry":       current.Industry(),
		"updated_at":     current.UpdatedAt(),
	}

	result := db.Table("org.organizations").Where("id = ?", id.UUID()).Updates(updates)
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

func (r *OrganizationRepo) CreateStorageObject(ctx context.Context, object appports.StorageObject) error {
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

func (r *OrganizationRepo) organizationLogoExists(db *gorm.DB, organizationID, objectID uuid.UUID) (bool, error) {
	var n int64
	err := db.Table("storage.objects").
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", objectID, organizationID).
		Limit(1).
		Count(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
