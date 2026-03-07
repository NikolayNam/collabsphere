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

func (h *Handler) Handle(ctx context.Context, orgID orgDomain.OrganizationID, accountID string, role string) error {
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

    normalizedRole := memberDomain.MembershipRole(strings.TrimSpace(role))
    if normalizedRole == "" {
        normalizedRole = memberDomain.MembershipRoleMember
    }
    if !normalizedRole.IsValid() {
        return errors.InvalidInput("Invalid role")
    }

    m, err := memberDomain.NewMembership(memberDomain.NewMembershipParams{
        OrganizationID: orgID,
        AccountID:      accID,
        Role:           normalizedRole,
        Now:            h.clock.Now(),
    })
    if err != nil {
        return errors.InvalidInput("Invalid membership")
    }

    exists, err := h.orgReader.Exists(ctx, orgID)
    if err != nil {
        return err
    }
    if !exists {
        return errors.OrganizationNotFound()
    }

    if err := h.repo.AddMember(ctx, orgID, m); err != nil {
        return err
    }

    return nil
}
