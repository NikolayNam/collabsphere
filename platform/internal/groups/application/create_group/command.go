package create_group

import (
	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
)

type Command struct {
	Name           string
	Slug           string
	Description    *string
	OwnerAccountID accdomain.AccountID
}
