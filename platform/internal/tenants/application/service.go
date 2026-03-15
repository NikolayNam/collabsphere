package application

import (
	"context"
	"strings"
	"time"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/tenants/application/ports"
	"github.com/NikolayNam/collabsphere/internal/tenants/domain"
	sharedtx "github.com/NikolayNam/collabsphere/shared/tx"
	"github.com/google/uuid"
)

type Clock interface {
	Now() time.Time
}

type Service struct {
	repo ports.TenantRepository
	tx   sharedtx.Manager
	clk  Clock
}

func New(repo ports.TenantRepository, tx sharedtx.Manager, clk Clock) *Service {
	return &Service{repo: repo, tx: tx, clk: clk}
}

func (s *Service) CreateTenant(ctx context.Context, actorAccountID uuid.UUID, name, slug string, description *string) (*domain.Tenant, error) {
	if actorAccountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required")
	}
	name = strings.TrimSpace(name)
	slug = domain.NormalizeSlug(slug)
	if name == "" || slug == "" {
		return nil, fault.Validation("Tenant name and slug are required")
	}

	now := s.clk.Now()
	tenant := &domain.Tenant{
		ID:          uuid.New(),
		Name:        name,
		Slug:        slug,
		Description: domain.NormalizeOptional(description),
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   &now,
	}

	member := &domain.TenantMember{
		ID:        uuid.New(),
		TenantID:  tenant.ID,
		AccountID: actorAccountID,
		Role:      domain.TenantRoleOwner,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: &now,
	}

	if err := s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.CreateTenant(ctx, tenant); err != nil {
			return err
		}
		return s.repo.AddTenantMember(ctx, member)
	}); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *Service) ListMyTenants(ctx context.Context, actorAccountID uuid.UUID) ([]domain.TenantMembershipView, error) {
	if actorAccountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required")
	}
	return s.repo.ListTenantsByAccount(ctx, actorAccountID)
}

func (s *Service) GetTenant(ctx context.Context, actorAccountID, tenantID uuid.UUID) (*domain.Tenant, error) {
	if _, err := s.requireTenantAccess(ctx, tenantID, actorAccountID, false); err != nil {
		return nil, err
	}
	tenant, err := s.repo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, fault.NotFound("Tenant not found")
	}
	return tenant, nil
}

func (s *Service) AddTenantMember(ctx context.Context, actorAccountID, tenantID, accountID uuid.UUID, role string) error {
	_, err := s.requireTenantAccess(ctx, tenantID, actorAccountID, true)
	if err != nil {
		return err
	}
	parsedRole := domain.TenantRole(strings.TrimSpace(strings.ToLower(role)))
	if parsedRole == "" {
		parsedRole = domain.TenantRoleMember
	}
	if !parsedRole.IsValid() {
		return fault.Validation("Invalid tenant role")
	}
	now := s.clk.Now()
	return s.repo.AddTenantMember(ctx, &domain.TenantMember{
		ID:        uuid.New(),
		TenantID:  tenantID,
		AccountID: accountID,
		Role:      parsedRole,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: &now,
	})
}

func (s *Service) ListTenantMembers(ctx context.Context, actorAccountID, tenantID uuid.UUID) ([]domain.TenantMember, error) {
	if _, err := s.requireTenantAccess(ctx, tenantID, actorAccountID, false); err != nil {
		return nil, err
	}
	return s.repo.ListTenantMembers(ctx, tenantID)
}

func (s *Service) AddTenantOrganization(ctx context.Context, actorAccountID, tenantID, organizationID uuid.UUID) error {
	_, err := s.requireTenantAccess(ctx, tenantID, actorAccountID, true)
	if err != nil {
		return err
	}
	now := s.clk.Now()
	return s.repo.AddTenantOrganization(ctx, &domain.TenantOrganization{
		ID:             uuid.New(),
		TenantID:       tenantID,
		OrganizationID: organizationID,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      &now,
	})
}

func (s *Service) ListTenantOrganizations(ctx context.Context, actorAccountID, tenantID uuid.UUID) ([]domain.TenantOrganization, error) {
	if _, err := s.requireTenantAccess(ctx, tenantID, actorAccountID, false); err != nil {
		return nil, err
	}
	return s.repo.ListTenantOrganizations(ctx, tenantID)
}

func (s *Service) requireTenantAccess(ctx context.Context, tenantID, actorAccountID uuid.UUID, manage bool) (*domain.TenantMember, error) {
	if tenantID == uuid.Nil {
		return nil, fault.Validation("Invalid tenant id")
	}
	if actorAccountID == uuid.Nil {
		return nil, fault.Unauthorized("Authentication required")
	}
	member, err := s.repo.GetTenantMember(ctx, tenantID, actorAccountID)
	if err != nil {
		return nil, err
	}
	if member == nil || !member.IsActive || member.DeletedAt != nil {
		return nil, fault.Forbidden("Tenant access denied")
	}
	if manage && !member.Role.CanManage() {
		return nil, fault.Forbidden("Only tenant owners or admins can manage tenant relations")
	}
	return member, nil
}
