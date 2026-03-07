package add_organization_member

import (
	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type Command struct {
	GroupID        domain.GroupID
	ActorAccountID accdomain.AccountID
	OrganizationID orgdomain.OrganizationID
}
