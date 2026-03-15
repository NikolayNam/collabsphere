package postgres

import (
	"context"
	"errors"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/tenants/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func (r *TenantRepo) CreateTenant(ctx context.Context, tenant *domain.Tenant) error {
	record := map[string]any{
		"id":          tenant.ID,
		"name":        tenant.Name,
		"slug":        tenant.Slug,
		"description": tenant.Description,
		"is_active":   tenant.IsActive,
		"created_at":  tenant.CreatedAt,
		"updated_at":  tenant.UpdatedAt,
	}
	if err := r.dbFrom(ctx).Table("tenant.tenants").Create(record).Error; err != nil {
		return mapTenantDBErr(err)
	}
	return nil
}

func (r *TenantRepo) AddTenantMember(ctx context.Context, member *domain.TenantMember) error {
	record := map[string]any{
		"id":         member.ID,
		"tenant_id":  member.TenantID,
		"account_id": member.AccountID,
		"role":       string(member.Role),
		"is_active":  member.IsActive,
		"created_at": member.CreatedAt,
		"updated_at": member.UpdatedAt,
		"deleted_at": member.DeletedAt,
	}
	if err := r.dbFrom(ctx).Table("tenant.tenant_account_members").Create(record).Error; err != nil {
		return mapTenantDBErr(err)
	}
	return nil
}

func (r *TenantRepo) AddTenantOrganization(ctx context.Context, rel *domain.TenantOrganization) error {
	record := map[string]any{
		"id":              rel.ID,
		"tenant_id":       rel.TenantID,
		"organization_id": rel.OrganizationID,
		"is_active":       rel.IsActive,
		"created_at":      rel.CreatedAt,
		"updated_at":      rel.UpdatedAt,
		"deleted_at":      rel.DeletedAt,
	}
	if err := r.dbFrom(ctx).Table("tenant.tenant_organizations").Create(record).Error; err != nil {
		return mapTenantDBErr(err)
	}
	return nil
}

func mapTenantDBErr(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		if errorsIsUniqueViolation(err) {
			return fault.Conflict("Entity already exists")
		}
		return err
	}
	switch pgErr.Code {
	case "23505":
		return fault.Conflict("Entity already exists")
	case "23503":
		return fault.Validation("Related entity not found")
	default:
		return err
	}
}

func errorsIsUniqueViolation(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
