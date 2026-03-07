package create_with_owner

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	groupsErrors "github.com/NikolayNam/collabsphere/internal/groups/application/errors"
	"github.com/NikolayNam/collabsphere/internal/groups/application/ports"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
	sharedtx "github.com/NikolayNam/collabsphere/shared/tx"
)

type Handler struct {
	tx   sharedtx.Manager
	repo ports.GroupRepository
}

func NewHandler(txm sharedtx.Manager, repo ports.GroupRepository) *Handler {
	return &Handler{tx: txm, repo: repo}
}

func (h *Handler) Handle(ctx context.Context, group *domain.Group, ownerAccountID accdomain.AccountID) error {
	if group == nil {
		return groupsErrors.InvalidInput("Group is required")
	}
	if ownerAccountID.IsZero() {
		return groupsErrors.InvalidInput("Owner account is required")
	}

	return h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := h.repo.Create(ctx, group); err != nil {
			return err
		}

		ownerMembership, err := domain.NewAccountMember(domain.NewAccountMemberParams{
			GroupID:   group.ID(),
			AccountID: ownerAccountID,
			Role:      domain.GroupAccountRoleOwner,
			Now:       group.CreatedAt(),
		})
		if err != nil {
			return groupsErrors.InvalidInput("Invalid group owner")
		}

		if err := h.repo.AddAccountMember(ctx, group.ID(), ownerMembership); err != nil {
			return err
		}

		return nil
	})
}
