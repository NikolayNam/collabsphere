package ports

import (
	"context"

	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type MembershipRepository interface {
	AddMember(ctx context.Context, orgID orgDomain.OrganizationID, m *memberDomain.Membership) error
	ListMembers(ctx context.Context, orgID orgDomain.OrganizationID) ([]memberDomain.MemberView, error)
}
