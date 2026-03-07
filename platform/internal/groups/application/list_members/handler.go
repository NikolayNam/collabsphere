package list_members

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

func (h *Handler) Handle(ctx context.Context, q Query) (*domain.MembersView, error) {
	if q.GroupID.IsZero() {
		return nil, groupsErrors.InvalidInput("Invalid group id")
	}
	if q.ActorAccountID.IsZero() {
		return nil, groupsErrors.AccessDenied()
	}

	group, err := h.repo.GetByID(ctx, q.GroupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, groupsErrors.GroupNotFound()
	}

	actorMembership, err := h.repo.GetAccountMember(ctx, q.GroupID, q.ActorAccountID)
	if err != nil {
		return nil, err
	}
	if actorMembership == nil || !actorMembership.IsActive() {
		return nil, groupsErrors.AccessDenied()
	}

	members, err := h.repo.ListMembers(ctx, q.GroupID)
	if err != nil {
		return nil, err
	}
	if members == nil {
		return &domain.MembersView{}, nil
	}

	return members, nil
}
