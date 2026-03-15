package get_group_by_id

import (
	"context"

	groupsErrors "github.com/NikolayNam/collabsphere/internal/groups/application/errors"
	"github.com/NikolayNam/collabsphere/internal/groups/application/ports"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
)

type Handler struct {
	repo ports.GroupRepository
}

func NewHandler(repo ports.GroupRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*domain.Group, error) {
	if q.ID.IsZero() {
		return nil, groupsErrors.InvalidInput("Invalid group id")
	}
	if q.ActorAccountID.IsZero() {
		return nil, groupsErrors.AccessDenied()
	}

	group, err := h.repo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, groupsErrors.GroupNotFound()
	}

	hasAccess, err := h.repo.HasGroupAccessForAccount(ctx, q.ID, q.ActorAccountID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, groupsErrors.AccessDenied()
	}

	return group, nil
}
