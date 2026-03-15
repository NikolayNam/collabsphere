package ports

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/tenants/domain"
	"github.com/google/uuid"
)

type TenantRepository interface {
	CreateTenant(ctx context.Context, tenant *domain.Tenant) error
	GetTenantByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error)
	ListTenantsByAccount(ctx context.Context, accountID uuid.UUID) ([]domain.TenantMembershipView, error)

	GetTenantMember(ctx context.Context, tenantID, accountID uuid.UUID) (*domain.TenantMember, error)
	AddTenantMember(ctx context.Context, member *domain.TenantMember) error
	ListTenantMembers(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantMember, error)

	AddTenantOrganization(ctx context.Context, rel *domain.TenantOrganization) error
	ListTenantOrganizations(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantOrganization, error)
}
