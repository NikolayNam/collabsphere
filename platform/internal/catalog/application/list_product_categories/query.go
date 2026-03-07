package list_product_categories

import (
	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type Query struct {
	OrganizationID orgdomain.OrganizationID
	ActorAccountID accdomain.AccountID
}
