package ports

import (
    "context"

    orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type MembershipWriter interface {
    AddMember(ctx context.Context, orgID orgDomain.OrganizationID, accountID string, role string) error
}
