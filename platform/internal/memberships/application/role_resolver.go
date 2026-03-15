package application

import (
	"context"
	"strings"

	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	"github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

// RoleResolverAdapter adapts OrganizationRoleRepository to RoleResolver port.
type RoleResolverAdapter struct {
	repo ports.OrganizationRoleRepository
}

func NewRoleResolverAdapter(repo ports.OrganizationRoleRepository) *RoleResolverAdapter {
	return &RoleResolverAdapter{repo: repo}
}

func (r *RoleResolverAdapter) ResolveRoleForPermissions(ctx context.Context, orgID orgDomain.OrganizationID, roleCode string) (memberDomain.MembershipRole, error) {
	role := memberDomain.MembershipRole(strings.ToLower(strings.TrimSpace(roleCode)))
	if role.IsValid() {
		return role, nil
	}
	custom, err := r.repo.GetByCode(ctx, orgID, roleCode)
	if err != nil {
		return "", err
	}
	if custom == nil || custom.IsDeleted() {
		return "", nil
	}
	return custom.BaseRole(), nil
}
