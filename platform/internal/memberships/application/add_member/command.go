package add_member

import (
    accountsDomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
    memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
    organizationDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type Command struct {
    OrganizationID organizationDomain.OrganizationID
    AccountID      accountsDomain.AccountID
    Role           memberDomain.MembershipRole
}
