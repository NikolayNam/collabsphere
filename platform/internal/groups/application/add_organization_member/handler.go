package add_organization_member

import (
	"context"

	groupsErrors "github.com/NikolayNam/collabsphere/internal/groups/application/errors"
	"github.com/NikolayNam/collabsphere/internal/groups/application/ports"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
)

type Handler struct {
	repo          ports.GroupRepository
	organizations ports.OrganizationReader
	clock         ports.Clock
}

func NewHandler(repo ports.GroupRepository, organizations ports.OrganizationReader, clock ports.Clock) *Handler {
	return &Handler{repo: repo, organizations: organizations, clock: clock}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	if cmd.GroupID.IsZero() || cmd.ActorAccountID.IsZero() || cmd.OrganizationID.IsZero() {
		return groupsErrors.InvalidInput("Invalid group or organization id")
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

	organization, err := h.organizations.GetByID(ctx, cmd.OrganizationID)
	if err != nil {
		return err
	}
	if organization == nil {
		return groupsErrors.InvalidInput("Organization not found")
	}

	member, err := domain.NewOrganizationMember(domain.NewOrganizationMemberParams{
		GroupID:        cmd.GroupID,
		OrganizationID: cmd.OrganizationID,
		Now:            h.clock.Now(),
	})
	if err != nil {
		return groupsErrors.InvalidInput("Invalid group organization member")
	}

	return h.repo.AddOrganizationMember(ctx, cmd.GroupID, member)
}
