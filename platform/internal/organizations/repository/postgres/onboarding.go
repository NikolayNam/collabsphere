package postgres

import (
	"context"
	"errors"
	"fmt"

	apperrors "github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/mapper"
	"gorm.io/gorm"
)

func (r *OrganizationRepo) GetCooperationApplication(ctx context.Context, organizationID domain.OrganizationID) (*domain.CooperationApplication, error) {
	if organizationID.IsZero() {
		return nil, apperrors.InvalidInput("Organization ID is required")
	}

	var model dbmodel.CooperationApplication
	err := r.dbFrom(ctx).WithContext(ctx).
		Where("organization_id = ?", organizationID.UUID()).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToDomainCooperationApplication(&model)
}

func (r *OrganizationRepo) SaveCooperationApplication(ctx context.Context, application *domain.CooperationApplication) (*domain.CooperationApplication, error) {
	if application == nil {
		return nil, apperrors.InvalidInput("Cooperation application is required")
	}
	model, err := mapper.ToDBCooperationApplication(application)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", apperrors.ErrValidation, err)
	}

	db := r.dbFrom(ctx).WithContext(ctx)
	if model.PriceListObjectID != nil {
		valid, err := r.organizationObjectExists(db, model.OrganizationID, *model.PriceListObjectID)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, apperrors.InvalidInput("Price list object does not exist or does not belong to organization")
		}
	}

	updates := map[string]any{
		"status":                 model.Status,
		"confirmation_email":     model.ConfirmationEmail,
		"company_name":           model.CompanyName,
		"represented_categories": model.RepresentedCategories,
		"minimum_order_amount":   model.MinimumOrderAmount,
		"delivery_geography":     model.DeliveryGeography,
		"sales_channels":         model.SalesChannels,
		"storefront_url":         model.StorefrontURL,
		"contact_first_name":     model.ContactFirstName,
		"contact_last_name":      model.ContactLastName,
		"contact_job_title":      model.ContactJobTitle,
		"price_list_object_id":   model.PriceListObjectID,
		"contact_email":          model.ContactEmail,
		"contact_phone":          model.ContactPhone,
		"partner_code":           model.PartnerCode,
		"review_note":            model.ReviewNote,
		"reviewer_account_id":    model.ReviewerAccountID,
		"submitted_at":           model.SubmittedAt,
		"reviewed_at":            model.ReviewedAt,
		"updated_at":             model.UpdatedAt,
	}

	var count int64
	if err := db.Table("org.cooperation_applications").Where("organization_id = ?", model.OrganizationID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		if err := db.Create(model).Error; err != nil {
			if isUniqueViolation(err) {
				return nil, fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
			}
			if isForeignKeyViolation(err) {
				return nil, apperrors.InvalidInput("Cooperation application references invalid entities")
			}
			return nil, err
		}
	} else {
		result := db.Table("org.cooperation_applications").Where("organization_id = ?", model.OrganizationID).Updates(updates)
		if result.Error != nil {
			if isUniqueViolation(result.Error) {
				return nil, fmt.Errorf("%w: %w", apperrors.ErrConflict, result.Error)
			}
			if isForeignKeyViolation(result.Error) {
				return nil, apperrors.InvalidInput("Cooperation application references invalid entities")
			}
			return nil, result.Error
		}
	}

	return r.GetCooperationApplication(ctx, application.OrganizationID())
}

func (r *OrganizationRepo) CreateOrganizationLegalDocument(ctx context.Context, document *domain.OrganizationLegalDocument) (*domain.OrganizationLegalDocument, error) {
	if document == nil {
		return nil, apperrors.InvalidInput("Organization legal document is required")
	}
	model := mapper.ToDBOrganizationLegalDocument(document)
	db := r.dbFrom(ctx).WithContext(ctx)
	valid, err := r.organizationObjectExists(db, model.OrganizationID, model.ObjectID)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, apperrors.InvalidInput("Legal document object does not exist or does not belong to organization")
	}
	if err := db.Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
		}
		if isForeignKeyViolation(err) {
			return nil, apperrors.InvalidInput("Organization legal document references invalid entities")
		}
		return nil, err
	}
	return mapper.ToDomainOrganizationLegalDocument(model)
}

func (r *OrganizationRepo) ListOrganizationLegalDocuments(ctx context.Context, organizationID domain.OrganizationID) ([]domain.OrganizationLegalDocument, error) {
	if organizationID.IsZero() {
		return nil, apperrors.InvalidInput("Organization ID is required")
	}
	var models []dbmodel.OrganizationLegalDocument
	err := r.dbFrom(ctx).WithContext(ctx).
		Where("organization_id = ? AND deleted_at IS NULL", organizationID.UUID()).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.OrganizationLegalDocument, 0, len(models))
	for i := range models {
		item, err := mapper.ToDomainOrganizationLegalDocument(&models[i])
		if err != nil {
			return nil, err
		}
		if item != nil {
			out = append(out, *item)
		}
	}
	return out, nil
}
