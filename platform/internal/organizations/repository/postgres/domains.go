package postgres

import (
	"context"
	"errors"
	"time"

	apperrors "github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/mapper"
	"gorm.io/gorm"
)

func (r *OrganizationRepo) ListDomains(ctx context.Context, organizationID domain.OrganizationID) ([]domain.OrganizationDomain, error) {
	if organizationID.IsZero() {
		return []domain.OrganizationDomain{}, nil
	}

	var models []dbmodel.OrganizationDomain
	err := r.dbFrom(ctx).WithContext(ctx).
		Where("organization_id = ? AND disabled_at IS NULL", organizationID.UUID()).
		Order("is_primary DESC, hostname ASC, created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	out := make([]domain.OrganizationDomain, 0, len(models))
	for _, model := range models {
		item, err := mapper.ToDomainOrganizationDomain(&model)
		if err != nil {
			return nil, err
		}
		if item != nil {
			out = append(out, *item)
		}
	}
	return out, nil
}

func (r *OrganizationRepo) ReplaceDomains(ctx context.Context, organizationID domain.OrganizationID, domains []domain.OrganizationDomain, now time.Time) ([]domain.OrganizationDomain, error) {
	if organizationID.IsZero() {
		return nil, apperrors.InvalidInput("Organization ID is required")
	}
	if now.IsZero() {
		return nil, apperrors.InvalidInput("UpdatedAt is required")
	}

	db := r.dbFrom(ctx).WithContext(ctx)

	var existing []dbmodel.OrganizationDomain
	if err := db.Where("organization_id = ? AND disabled_at IS NULL", organizationID.UUID()).Find(&existing).Error; err != nil {
		return nil, err
	}
	existingByHostname := make(map[string]dbmodel.OrganizationDomain, len(existing))
	for _, item := range existing {
		existingByHostname[item.Hostname] = item
	}

	desired := make(map[string]domain.OrganizationDomain, len(domains))
	for _, item := range domains {
		desired[item.Hostname()] = item
		if current, ok := existingByHostname[item.Hostname()]; ok {
			updates := map[string]any{
				"kind":        string(item.Kind()),
				"is_primary":  item.IsPrimary(),
				"verified_at": item.VerifiedAt(),
				"disabled_at": nil,
				"updated_at":  now,
			}
			if err := db.Table("org.organization_domains").Where("id = ?", current.ID).Updates(updates).Error; err != nil {
				if isUniqueViolation(err) {
					return nil, apperrors.OrganizationDomainAlreadyExists()
				}
				return nil, err
			}
			continue
		}

		model := mapper.ToDBOrganizationDomainForCreate(item)
		if model == nil {
			return nil, apperrors.Internal("db organization domain model is nil", nil)
		}
		if err := db.Create(model).Error; err != nil {
			if isUniqueViolation(err) {
				return nil, apperrors.OrganizationDomainAlreadyExists()
			}
			return nil, err
		}
	}

	for _, item := range existing {
		if _, ok := desired[item.Hostname]; ok {
			continue
		}
		updates := map[string]any{
			"is_primary":  false,
			"disabled_at": now,
			"updated_at":  now,
		}
		if err := db.Table("org.organization_domains").Where("id = ?", item.ID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return r.ListDomains(ctx, organizationID)
}

func (r *OrganizationRepo) GetByHostname(ctx context.Context, hostname string) (*domain.Organization, error) {
	var model dbmodel.Organization
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("org.organizations AS o").
		Select("o.*").
		Joins("JOIN org.organization_domains AS d ON d.organization_id = o.id").
		Where("d.hostname = ? AND d.disabled_at IS NULL AND d.verified_at IS NOT NULL AND o.is_active = TRUE", hostname).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToDomainOrganization(&model)
}
