package create_organization

import (
	"context"
	stdErrors "errors"

	"github.com/NikolayNam/collabsphere/internal/organizations/application/create_organization_with_owner"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type Handler struct {
	creator *create_with_owner.Handler
	clock   ports.Clock
}

func NewHandler(creator *create_with_owner.Handler, clock ports.Clock) *Handler {
	return &Handler{creator: creator, clock: clock}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*domain.Organization, error) {
	if cmd.OwnerAccountID.IsZero() {
		return nil, errors.InvalidInput("Owner account is required")
	}

	organization, err := domain.NewOrganization(domain.NewOrganizationParams{
		ID:   domain.NewOrganizationID(),
		Name: cmd.Name,
		Slug: cmd.Slug,
		Now:  h.clock.Now(),
	})
	if err != nil {
		return nil, errors.InvalidInput("Invalid organization data")
	}

	if err := h.creator.Handle(ctx, organization, cmd.OwnerAccountID); err != nil {
		if stdErrors.Is(err, errors.ErrConflict) {
			return nil, errors.OrganizationAlreadyExists()
		}
		return nil, err
	}

	return organization, nil
}
