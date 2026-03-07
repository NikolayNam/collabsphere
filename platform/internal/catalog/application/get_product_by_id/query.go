package get_product_by_id

import (
	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type Query struct {
	OrganizationID orgdomain.OrganizationID
	ActorAccountID accdomain.AccountID
	ProductID      catalogdomain.ProductID
}
