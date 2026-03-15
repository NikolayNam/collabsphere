package ports

import (
	"context"

	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type OrganizationRoleRepository interface {
	Create(ctx context.Context, role *memberDomain.OrganizationRole) error
	Save(ctx context.Context, role *memberDomain.OrganizationRole) error
	GetByID(ctx context.Context, orgID orgDomain.OrganizationID, roleID uuid.UUID) (*memberDomain.OrganizationRole, error)
	GetByCode(ctx context.Context, orgID orgDomain.OrganizationID, code string) (*memberDomain.OrganizationRole, error)
	List(ctx context.Context, orgID orgDomain.OrganizationID, includeDeleted bool) ([]*memberDomain.OrganizationRole, error)
	CountMembersWithRole(ctx context.Context, orgID orgDomain.OrganizationID, roleCode string) (int64, error)
}
