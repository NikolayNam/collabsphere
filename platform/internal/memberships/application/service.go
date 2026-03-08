package application

import (
	"context"
	"strings"

	accDomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	memberErrors "github.com/NikolayNam/collabsphere/internal/memberships/application/errors"
	memberPorts "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgPorts "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/google/uuid"
)

var (
	ErrValidation = memberErrors.ErrValidation
)

type orgReaderAdapter struct {
	repo orgPorts.OrganizationRepository
}

func (a orgReaderAdapter) Exists(ctx context.Context, id orgDomain.OrganizationID) (bool, error) {
	org, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	return org != nil, nil
}

type UpdateMemberCmd struct {
	OrganizationID orgDomain.OrganizationID
	MembershipID   uuid.UUID
	ActorAccountID uuid.UUID
	Role           *string
	IsActive       *bool
}

type RemoveMemberCmd struct {
	OrganizationID orgDomain.OrganizationID
	MembershipID   uuid.UUID
	ActorAccountID uuid.UUID
}

type Service struct {
	repo      memberPorts.MembershipRepository
	orgReader orgReaderAdapter
	clock     memberPorts.Clock
}

func New(memberRepo memberPorts.MembershipRepository, orgRepo orgPorts.OrganizationRepository, clock memberPorts.Clock) *Service {
	return &Service{
		repo:      memberRepo,
		orgReader: orgReaderAdapter{repo: orgRepo},
		clock:     clock,
	}
}

func (s *Service) AddMember(ctx context.Context, actorAccountID uuid.UUID, orgID orgDomain.OrganizationID, accountID string, role string) (*memberDomain.MemberView, error) {
	actorMembership, err := s.requireManageableMembership(ctx, orgID, actorAccountID)
	if err != nil {
		return nil, err
	}

	targetAccountID, err := parseAccountID(accountID)
	if err != nil {
		return nil, memberErrors.InvalidInput("Invalid account_id")
	}
	targetRole := parseRole(role, memberDomain.MembershipRoleMember)
	if !targetRole.IsValid() {
		return nil, memberErrors.InvalidInput("Invalid role")
	}
	if !actorMembership.Role().CanAssign(targetRole) {
		return nil, fault.Forbidden("Membership role assignment is not allowed")
	}

	existing, err := s.repo.GetMemberByAccount(ctx, orgID, targetAccountID)
	if err != nil {
		return nil, memberErrors.Internal("get member by account", err)
	}

	now := s.clock.Now()
	if existing == nil {
		created, err := memberDomain.NewMembership(memberDomain.NewMembershipParams{
			OrganizationID: orgID,
			AccountID:      targetAccountID,
			Role:           targetRole,
			Now:            now,
		})
		if err != nil {
			return nil, memberErrors.InvalidInput("Invalid membership")
		}
		if err := s.repo.AddMember(ctx, orgID, created); err != nil {
			return nil, err
		}
	} else {
		if !actorMembership.Role().CanManageTarget(existing.Role()) {
			return nil, fault.Forbidden("Membership change is not allowed for the selected member")
		}
		if existing.IsActive() && !existing.IsRemoved() {
			return nil, memberErrors.MemberAlreadyExists()
		}
		if err := existing.ChangeRole(targetRole, now); err != nil {
			return nil, memberErrors.InvalidInput("Invalid role")
		}
		if err := existing.Activate(now); err != nil {
			return nil, memberErrors.InvalidInput("Invalid membership state")
		}
		if err := s.repo.SaveMember(ctx, orgID, existing); err != nil {
			return nil, memberErrors.Internal("save member", err)
		}
	}

	return s.getMemberViewByAccount(ctx, orgID, targetAccountID)
}

func (s *Service) UpdateMember(ctx context.Context, cmd UpdateMemberCmd) (*memberDomain.MemberView, error) {
	actorMembership, err := s.requireManageableMembership(ctx, cmd.OrganizationID, cmd.ActorAccountID)
	if err != nil {
		return nil, err
	}
	target, err := s.repo.GetMemberByID(ctx, cmd.OrganizationID, cmd.MembershipID)
	if err != nil {
		return nil, memberErrors.Internal("get member by id", err)
	}
	if target == nil {
		return nil, fault.NotFound("Membership not found")
	}
	if !actorMembership.Role().CanManageTarget(target.Role()) {
		return nil, fault.Forbidden("Membership change is not allowed for the selected member")
	}

	now := s.clock.Now()
	nextRole := target.Role()
	if cmd.Role != nil {
		nextRole = parseRole(*cmd.Role, target.Role())
		if !nextRole.IsValid() {
			return nil, memberErrors.InvalidInput("Invalid role")
		}
		if !actorMembership.Role().CanAssign(nextRole) {
			return nil, fault.Forbidden("Membership role assignment is not allowed")
		}
	}

	nextActive := target.IsActive()
	if cmd.IsActive != nil {
		nextActive = *cmd.IsActive
	}

	if target.Role() == memberDomain.MembershipRoleOwner && (!nextActive || nextRole != memberDomain.MembershipRoleOwner) {
		if err := s.ensureAnotherActiveOwner(ctx, cmd.OrganizationID); err != nil {
			return nil, err
		}
	}

	if nextRole != target.Role() {
		if err := target.ChangeRole(nextRole, now); err != nil {
			return nil, memberErrors.InvalidInput("Invalid role")
		}
	}

	if cmd.IsActive != nil {
		if *cmd.IsActive {
			if err := target.Activate(now); err != nil {
				return nil, memberErrors.InvalidInput("Invalid membership state")
			}
		} else {
			if err := target.Suspend(now); err != nil {
				return nil, memberErrors.InvalidInput("Invalid membership state")
			}
		}
	}

	if err := s.repo.SaveMember(ctx, cmd.OrganizationID, target); err != nil {
		return nil, memberErrors.Internal("save member", err)
	}
	return s.getMemberViewByID(ctx, cmd.OrganizationID, cmd.MembershipID)
}

func (s *Service) RemoveMember(ctx context.Context, cmd RemoveMemberCmd) error {
	actorMembership, err := s.requireManageableMembership(ctx, cmd.OrganizationID, cmd.ActorAccountID)
	if err != nil {
		return err
	}
	target, err := s.repo.GetMemberByID(ctx, cmd.OrganizationID, cmd.MembershipID)
	if err != nil {
		return memberErrors.Internal("get member by id", err)
	}
	if target == nil {
		return fault.NotFound("Membership not found")
	}
	if !actorMembership.Role().CanManageTarget(target.Role()) {
		return fault.Forbidden("Membership removal is not allowed for the selected member")
	}
	if target.Role() == memberDomain.MembershipRoleOwner && target.IsActive() {
		if err := s.ensureAnotherActiveOwner(ctx, cmd.OrganizationID); err != nil {
			return err
		}
	}
	if err := target.Remove(s.clock.Now()); err != nil {
		return memberErrors.InvalidInput("Invalid membership state")
	}
	if err := s.repo.SaveMember(ctx, cmd.OrganizationID, target); err != nil {
		return memberErrors.Internal("save member", err)
	}
	return nil
}

func (s *Service) ListMembers(ctx context.Context, actorAccountID uuid.UUID, orgID orgDomain.OrganizationID) ([]memberDomain.MemberView, error) {
	if _, err := s.requireReadableMembership(ctx, orgID, actorAccountID); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(ctx, orgID)
}

func (s *Service) requireReadableMembership(ctx context.Context, orgID orgDomain.OrganizationID, actorAccountID uuid.UUID) (*memberDomain.Membership, error) {
	if orgID.IsZero() {
		return nil, memberErrors.InvalidInput("Invalid organization_id")
	}
	actorID, err := accDomain.AccountIDFromUUID(actorAccountID)
	if err != nil || actorID.IsZero() {
		return nil, fault.Unauthorized("Authentication required")
	}
	exists, err := s.orgReader.Exists(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, memberErrors.OrganizationNotFound()
	}
	membership, err := s.repo.GetMemberByAccount(ctx, orgID, actorID)
	if err != nil {
		return nil, memberErrors.Internal("get actor membership", err)
	}
	if membership == nil || !membership.IsActive() || membership.IsRemoved() {
		return nil, fault.Forbidden("Organization access denied")
	}
	return membership, nil
}

func (s *Service) requireManageableMembership(ctx context.Context, orgID orgDomain.OrganizationID, actorAccountID uuid.UUID) (*memberDomain.Membership, error) {
	membership, err := s.requireReadableMembership(ctx, orgID, actorAccountID)
	if err != nil {
		return nil, err
	}
	if !membership.Role().CanManageMembers() {
		return nil, fault.Forbidden("Only organization owners or admins can manage members")
	}
	return membership, nil
}

func (s *Service) ensureAnotherActiveOwner(ctx context.Context, orgID orgDomain.OrganizationID) error {
	count, err := s.repo.CountActiveMembersByRole(ctx, orgID, memberDomain.MembershipRoleOwner)
	if err != nil {
		return memberErrors.Internal("count active owners", err)
	}
	if count <= 1 {
		return fault.Conflict("Organization must retain at least one active owner")
	}
	return nil
}

func (s *Service) getMemberViewByAccount(ctx context.Context, orgID orgDomain.OrganizationID, accountID accDomain.AccountID) (*memberDomain.MemberView, error) {
	members, err := s.repo.ListMembers(ctx, orgID)
	if err != nil {
		return nil, memberErrors.Internal("list members", err)
	}
	for i := range members {
		if members[i].AccountID == accountID.UUID() {
			return &members[i], nil
		}
	}
	return nil, fault.NotFound("Membership not found")
}

func (s *Service) getMemberViewByID(ctx context.Context, orgID orgDomain.OrganizationID, membershipID uuid.UUID) (*memberDomain.MemberView, error) {
	members, err := s.repo.ListMembers(ctx, orgID)
	if err != nil {
		return nil, memberErrors.Internal("list members", err)
	}
	for i := range members {
		if members[i].MembershipID == membershipID {
			return &members[i], nil
		}
	}
	return nil, fault.NotFound("Membership not found")
}

func parseAccountID(raw string) (accDomain.AccountID, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || parsed == uuid.Nil {
		return accDomain.AccountID{}, memberErrors.ErrValidation
	}
	return accDomain.AccountIDFromUUID(parsed)
}

func parseRole(raw string, fallback memberDomain.MembershipRole) memberDomain.MembershipRole {
	parsed := memberDomain.ParseMembershipRole(raw)
	if parsed == "" {
		return fallback
	}
	return parsed
}
