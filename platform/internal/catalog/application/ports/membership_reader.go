package ports

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type MembershipReader interface {
	GetMemberByAccount(ctx context.Context, orgID orgdomain.OrganizationID, accountID accdomain.AccountID) (*memberDomain.Membership, error)
}
