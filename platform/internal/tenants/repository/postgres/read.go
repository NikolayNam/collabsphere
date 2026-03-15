package postgres

import (
	"context"
	"time"

	"github.com/NikolayNam/collabsphere/internal/tenants/domain"
	"github.com/google/uuid"
)

type tenantRow struct {
	ID          uuid.UUID  `gorm:"column:id"`
	Name        string     `gorm:"column:name"`
	Slug        string     `gorm:"column:slug"`
	Description *string    `gorm:"column:description"`
	IsActive    bool       `gorm:"column:is_active"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at"`
}

func (tenantRow) TableName() string { return "tenant.tenants" }

type tenantMembershipRow struct {
	ID             uuid.UUID  `gorm:"column:id"`
	Name           string     `gorm:"column:name"`
	Slug           string     `gorm:"column:slug"`
	Description    *string    `gorm:"column:description"`
	IsActive       bool       `gorm:"column:is_active"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at"`
	MembershipRole string     `gorm:"column:membership_role"`
}

type tenantMemberRow struct {
	ID        uuid.UUID  `gorm:"column:id"`
	TenantID  uuid.UUID  `gorm:"column:tenant_id"`
	AccountID uuid.UUID  `gorm:"column:account_id"`
	Role      string     `gorm:"column:role"`
	IsActive  bool       `gorm:"column:is_active"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (tenantMemberRow) TableName() string { return "tenant.tenant_account_members" }

type tenantOrgRow struct {
	ID             uuid.UUID  `gorm:"column:id"`
	TenantID       uuid.UUID  `gorm:"column:tenant_id"`
	OrganizationID uuid.UUID  `gorm:"column:organization_id"`
	IsActive       bool       `gorm:"column:is_active"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (tenantOrgRow) TableName() string { return "tenant.tenant_organizations" }

func (r *TenantRepo) GetTenantByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	var row tenantRow
	err := r.dbFrom(ctx).Table("tenant.tenants").
		Where("id = ? AND is_active = TRUE", tenantID).
		Take(&row).Error
	if err != nil {
		if isRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.Tenant{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Description: row.Description,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *TenantRepo) ListTenantsByAccount(ctx context.Context, accountID uuid.UUID) ([]domain.TenantMembershipView, error) {
	var rows []tenantMembershipRow
	err := r.dbFrom(ctx).
		Table("tenant.tenants t").
		Select("t.id, t.name, t.slug, t.description, t.is_active, t.created_at, t.updated_at, tam.role AS membership_role").
		Joins("JOIN tenant.tenant_account_members tam ON tam.tenant_id = t.id").
		Where("tam.account_id = ? AND tam.is_active = TRUE AND tam.deleted_at IS NULL AND t.is_active = TRUE", accountID).
		Order("t.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.TenantMembershipView, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.TenantMembershipView{
			ID:             row.ID,
			Name:           row.Name,
			Slug:           row.Slug,
			Description:    row.Description,
			IsActive:       row.IsActive,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
			MembershipRole: domain.TenantRole(row.MembershipRole),
		})
	}
	return out, nil
}

func (r *TenantRepo) GetTenantMember(ctx context.Context, tenantID, accountID uuid.UUID) (*domain.TenantMember, error) {
	var row tenantMemberRow
	err := r.dbFrom(ctx).
		Table("tenant.tenant_account_members").
		Where("tenant_id = ? AND account_id = ? AND deleted_at IS NULL", tenantID, accountID).
		Take(&row).Error
	if err != nil {
		if isRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.TenantMember{
		ID:        row.ID,
		TenantID:  row.TenantID,
		AccountID: row.AccountID,
		Role:      domain.TenantRole(row.Role),
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: row.DeletedAt,
	}, nil
}

func (r *TenantRepo) ListTenantMembers(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantMember, error) {
	var rows []tenantMemberRow
	err := r.dbFrom(ctx).
		Table("tenant.tenant_account_members").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.TenantMember, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.TenantMember{
			ID:        row.ID,
			TenantID:  row.TenantID,
			AccountID: row.AccountID,
			Role:      domain.TenantRole(row.Role),
			IsActive:  row.IsActive,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			DeletedAt: row.DeletedAt,
		})
	}
	return out, nil
}

func (r *TenantRepo) ListTenantOrganizations(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantOrganization, error) {
	var rows []tenantOrgRow
	err := r.dbFrom(ctx).
		Table("tenant.tenant_organizations").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.TenantOrganization, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.TenantOrganization{
			ID:             row.ID,
			TenantID:       row.TenantID,
			OrganizationID: row.OrganizationID,
			IsActive:       row.IsActive,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
			DeletedAt:      row.DeletedAt,
		})
	}
	return out, nil
}
