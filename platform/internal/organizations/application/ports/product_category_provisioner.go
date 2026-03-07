package ports

import (
	"context"
	"time"

	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type ProductCategoryProvisioner interface {
	ProvisionDefaults(ctx context.Context, organizationID domain.OrganizationID, now time.Time) error
}
