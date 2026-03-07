package ports

import (
	"context"

	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type MembershipReader interface {
	ListMembers(ctx context.Context, orgID orgdomain.OrganizationID) ([]memberDomain.MemberView, error)
}
