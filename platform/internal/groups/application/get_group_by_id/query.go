package get_group_by_id

import (
	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
)

type Query struct {
	ID             domain.GroupID
	ActorAccountID accdomain.AccountID
}
