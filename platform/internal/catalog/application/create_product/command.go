package create_product

import (
	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type Command struct {
	OrganizationID orgdomain.OrganizationID
	ActorAccountID accdomain.AccountID
	CategoryID     *catalogdomain.ProductCategoryID
	Name           string
	Description    *string
	SKU            *string
	PriceAmount    *string
	CurrencyCode   *string
	IsActive       *bool
}
