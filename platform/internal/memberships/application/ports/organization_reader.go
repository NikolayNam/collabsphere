package ports

import (
	"context"

	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type OrganizationReader interface {
	Exists(ctx context.Context, id orgDomain.OrganizationID) (bool, error)
}
