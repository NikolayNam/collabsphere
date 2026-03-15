package postgres

import (
	"context"
	"errors"
	"time"

	appports "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"gorm.io/gorm"

	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres/mapper"
	"github.com/google/uuid"
)

func (r *OrganizationRepo) GetByID(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error) {
	var m dbmodel.Organization

	err := r.dbFrom(ctx).WithContext(ctx).
		Take(&m, "id = ?", id.UUID()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mapper.ToDomainOrganization(&m)
}

func (r *OrganizationRepo) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]appports.OrganizationMembershipView, error) {
	type row struct {
		ID             uuid.UUID
		Name           string
		Slug           string
		LogoObjectID   *uuid.UUID
		IsActive       bool
		CreatedAt      time.Time
		UpdatedAt      *time.Time
		MembershipRole string
	}

	var rows []row
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("org.organizations AS o").
		Select("o.id, o.name, o.slug, o.logo_object_id, o.is_active, o.created_at, o.updated_at, m.role AS membership_role").
		Joins("JOIN iam.memberships AS m ON m.organization_id = o.id").
		Where("m.account_id = ? AND m.is_active = ? AND m.deleted_at IS NULL", accountID, true).
		Order("o.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]appports.OrganizationMembershipView, 0, len(rows))
	for _, row := range rows {
		var updatedAt *time.Time
		if row.UpdatedAt != nil {
			value := *row.UpdatedAt
			updatedAt = &value
		}
		out = append(out, appports.OrganizationMembershipView{
			ID:             row.ID,
			Name:           row.Name,
			Slug:           row.Slug,
			LogoObjectID:   row.LogoObjectID,
			IsActive:       row.IsActive,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      updatedAt,
			MembershipRole: row.MembershipRole,
		})
	}
	return out, nil
}

func (r *OrganizationRepo) ListActiveOrganizations(ctx context.Context, limit int) ([]appports.OrganizationListItem, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	var rows []appports.OrganizationListItem
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("org.organizations").
		Select("id, name, slug").
		Where("is_active = ?", true).
		Order("updated_at DESC NULLS LAST, created_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *OrganizationRepo) Exists(ctx context.Context, id domain.OrganizationID) (bool, error) {
	if id.IsZero() {
		return false, nil
	}

	var n int64
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("org.organizations").
		Where("id = ?", id.UUID()).
		Limit(1).
		Count(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
