package add_member

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/NikolayNam/collabsphere/internal/memberships/application/errors"
	"github.com/NikolayNam/collabsphere/internal/memberships/application/ports"

	acc "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type Handler struct {
	repo      ports.MembershipRepository
	orgReader ports.OrganizationReader
	clock     ports.Clock
}

func NewHandler(repo ports.MembershipRepository, orgReader ports.OrganizationReader, clock ports.Clock) *Handler {
	return &Handler{repo: repo, orgReader: orgReader, clock: clock}
}

// Handle adds a member to organization.
// Inputs kept close to delivery layer to avoid duplicating parsing there.
func (h *Handler) Handle(ctx context.Context, orgID orgDomain.OrganizationID, accountID string, kind string) error {
	if orgID.IsZero() {
		return errors.InvalidInput("Invalid organization_id")
	}

	rawAcc := strings.TrimSpace(accountID)
	uid, err := uuid.Parse(rawAcc)
	if err != nil || uid == uuid.Nil {
		return errors.InvalidInput("Invalid account_id")
	}

	accID, err := acc.AccountIDFromUUID(uid)
	if err != nil || accID.IsZero() {
		return errors.InvalidInput("Invalid account_id")
	}

	k := memberDomain.MembershipKind(strings.TrimSpace(kind))
	if k == "" {
		k = memberDomain.MembershipKindMember
	}
	if !k.IsValid() {
		return errors.InvalidInput("Invalid kind")
	}

	m, err := memberDomain.NewMembership(memberDomain.NewMembershipParams{
		OrganizationID: orgID,
		AccountID:      accID,
		Kind:           k,
		Status:         memberDomain.MembershipStatusActive,
		Now:            h.clock.Now(),
	})
	if err != nil {
		return errors.InvalidInput("Invalid membership")
	}

	// Optional: ensure organization exists to return NotFound instead of FK/InvalidInput.
	exists, err := h.orgReader.Exists(ctx, orgID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.OrganizationNotFound()
	}

	if err := h.repo.AddMember(ctx, orgID, m); err != nil {
		// Repo should map unique violation -> errors.MemberAlreadyExists().
		return err
	}

	return nil
}
