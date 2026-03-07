package ports

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type OrganizationWriter interface {
	Create(ctx context.Context, t *domain.Organization) error
}

type OrganizationReader interface {
	GetByID(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error)
}
