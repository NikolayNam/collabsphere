package create_organization

import accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"

type Command struct {
	Name           string
	Slug           string
	OwnerAccountID accdomain.AccountID
}
