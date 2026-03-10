package create_organization

import (
	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type Command struct {
	Name           string
	Slug           string
	OwnerAccountID accdomain.AccountID
	Domains        []orgdomain.OrganizationDomainDraft
}
