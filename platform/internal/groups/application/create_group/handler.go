package create_group

import (
	"context"
	stdErrors "errors"

	create_with_owner "github.com/NikolayNam/collabsphere/internal/groups/application/create_group_with_owner"
	groupsErrors "github.com/NikolayNam/collabsphere/internal/groups/application/errors"
	"github.com/NikolayNam/collabsphere/internal/groups/application/ports"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
)

type Handler struct {
	creator *create_with_owner.Handler
	clock   ports.Clock
}

func NewHandler(creator *create_with_owner.Handler, clock ports.Clock) *Handler {
	return &Handler{creator: creator, clock: clock}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*domain.Group, error) {
	if cmd.OwnerAccountID.IsZero() {
		return nil, groupsErrors.InvalidInput("Owner account is required")
	}

	group, err := domain.NewGroup(domain.NewGroupParams{
		ID:          domain.NewGroupID(),
		Name:        cmd.Name,
		Slug:        cmd.Slug,
		Description: cmd.Description,
		Now:         h.clock.Now(),
	})
	if err != nil {
		return nil, groupsErrors.InvalidInput("Invalid group data")
	}

	if err := h.creator.Handle(ctx, group, cmd.OwnerAccountID); err != nil {
		if stdErrors.Is(err, groupsErrors.ErrConflict) {
			return nil, groupsErrors.GroupAlreadyExists()
		}
		return nil, err
	}

	return group, nil
}
