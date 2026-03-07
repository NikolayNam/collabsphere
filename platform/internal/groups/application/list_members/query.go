package list_members

import (
	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
)

type Query struct {
	GroupID        domain.GroupID
	ActorAccountID accdomain.AccountID
}
