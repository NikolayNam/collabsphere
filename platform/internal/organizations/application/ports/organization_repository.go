package ports

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type OrganizationRepository interface {
	Create(ctx context.Context, t *domain.Organization) error
	GetByID(ctx context.Context, id domain.OrganizationID) (*domain.Organization, error)
}
