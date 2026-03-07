package run_product_import

import (
	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type Command struct {
	OrganizationID orgdomain.OrganizationID
	ActorAccountID accdomain.AccountID
	SourceObjectID uuid.UUID
	Mode           *string
}
