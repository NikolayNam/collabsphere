package add_account_member

import (
	"context"
	"strings"

	groupsErrors "github.com/NikolayNam/collabsphere/internal/groups/application/errors"
	"github.com/NikolayNam/collabsphere/internal/groups/application/ports"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
)

type Handler struct {
	repo     ports.GroupRepository
	accounts ports.AccountReader
	clock    ports.Clock
}

func NewHandler(repo ports.GroupRepository, accounts ports.AccountReader, clock ports.Clock) *Handler {
	return &Handler{repo: repo, accounts: accounts, clock: clock}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	if cmd.GroupID.IsZero() || cmd.ActorAccountID.IsZero() || cmd.AccountID.IsZero() {
		return groupsErrors.InvalidInput("Invalid group or account id")
	}

	group, err := h.repo.GetByID(ctx, cmd.GroupID)
	if err != nil {
		return err
	}
	if group == nil {
		return groupsErrors.GroupNotFound()
	}

	actorMembership, err := h.repo.GetAccountMember(ctx, cmd.GroupID, cmd.ActorAccountID)
	if err != nil {
		return err
	}
	if actorMembership == nil || !actorMembership.IsActive() || actorMembership.Role() != domain.GroupAccountRoleOwner {
		return groupsErrors.AccessDenied()
	}

	account, err := h.accounts.GetByID(ctx, cmd.AccountID)
	if err != nil {
		return err
	}
	if account == nil {
		return groupsErrors.InvalidInput("Account not found")
	}

	role := domain.GroupAccountRole(strings.TrimSpace(cmd.Role))
	if role == "" {
		role = domain.GroupAccountRoleMember
	}
	if !role.IsValid() {
		return groupsErrors.InvalidInput("Invalid group account role")
	}

	member, err := domain.NewAccountMember(domain.NewAccountMemberParams{
		GroupID:   cmd.GroupID,
		AccountID: cmd.AccountID,
		Role:      role,
		Now:       h.clock.Now(),
	})
	if err != nil {
		return groupsErrors.InvalidInput("Invalid group account member")
	}

	return h.repo.AddAccountMember(ctx, cmd.GroupID, member)
}
