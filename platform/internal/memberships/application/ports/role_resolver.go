package ports

import (
	"context"

	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

// RoleResolver resolves a role code to its base role for permission checks.
// System roles resolve to themselves; custom roles resolve to their base_role.
type RoleResolver interface {
	ResolveRoleForPermissions(ctx context.Context, orgID orgDomain.OrganizationID, roleCode string) (memberDomain.MembershipRole, error)
}
