package add_account_member

import (
	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
)

type Command struct {
	GroupID        domain.GroupID
	ActorAccountID accdomain.AccountID
	AccountID      accdomain.AccountID
	Role           string
}
