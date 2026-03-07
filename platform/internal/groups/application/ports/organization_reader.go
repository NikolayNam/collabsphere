package ports

import (
	"context"

	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type OrganizationReader interface {
	GetByID(ctx context.Context, id orgdomain.OrganizationID) (*orgdomain.Organization, error)
}
