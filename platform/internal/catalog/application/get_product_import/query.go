package get_product_import

import (
	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type Query struct {
	OrganizationID orgdomain.OrganizationID
	ActorAccountID accdomain.AccountID
	BatchID        uuid.UUID
}
