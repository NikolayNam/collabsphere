package application

import (
	"context"
	"strings"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	accesspolicy "github.com/NikolayNam/collabsphere/internal/iam/access"
	memberErrors "github.com/NikolayNam/collabsphere/internal/memberships/application/errors"
	memberPorts "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgPorts "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/requestctx"
	sharedtx "github.com/NikolayNam/collabsphere/shared/tx"
	"github.com/google/uuid"
)

const defaultOrganizationInvitationTTL = 7 * 24 * time.Hour

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

type CreateInvitationCmd struct {
	OrganizationID orgDomain.OrganizationID
	ActorAccountID uuid.UUID
	Email          string
	Role           string
}

type CreateInvitationResult struct {
	Invitation memberDomain.InvitationView
	Token      string
}

type AcceptInvitationCmd struct {
	Token          string
	ActorAccountID uuid.UUID
}

type AcceptInvitationResult struct {
	Invitation memberDomain.InvitationView
	Member     memberDomain.MemberView
}

type CreateAccessRequestCmd struct {
	OrganizationID orgDomain.OrganizationID
	ActorAccountID uuid.UUID
	Role           string
	Message        *string
}

type ReviewAccessRequestCmd struct {
	OrganizationID orgDomain.OrganizationID
	RequestID      uuid.UUID
	ActorAccountID uuid.UUID
	Decision       string
	ReviewNote     *string
}

type AccessRequestReviewDecision string

const (
	AccessRequestDecisionApprove AccessRequestReviewDecision = "approve"
	AccessRequestDecisionReject  AccessRequestReviewDecision = "reject"
)

type ListOrganizationRolesQuery struct {
	OrganizationID   orgDomain.OrganizationID
	ActorAccountID  uuid.UUID
	IncludeDeleted  bool
}

type CreateOrganizationRoleCmd struct {
	OrganizationID  orgDomain.OrganizationID
	ActorAccountID  uuid.UUID
	Code            string
	Name            string
	Description     string
	BaseRole        string
}

type UpdateOrganizationRoleCmd struct {
	OrganizationID  orgDomain.OrganizationID
	RoleID          uuid.UUID
	ActorAccountID  uuid.UUID
	Name            *string
	Description     *string
	BaseRole        *string
}

type DeleteOrganizationRoleCmd struct {
	OrganizationID  orgDomain.OrganizationID
	RoleID          uuid.UUID
	ActorAccountID  uuid.UUID
}

type Service struct {
	repo           memberPorts.MembershipRepository
	roleRepo       memberPorts.OrganizationRoleRepository
	accounts       memberPorts.AccountReader
	orgReader      orgReaderAdapter
	tx             sharedtx.Manager
	tokenGenerator memberPorts.TokenGenerator
	audit          memberPorts.AccessAuditRepository
	clock          memberPorts.Clock
	invitationTTL  time.Duration
}

func New(memberRepo memberPorts.MembershipRepository, roleRepo memberPorts.OrganizationRoleRepository, orgRepo orgPorts.OrganizationRepository, accountRepo memberPorts.AccountReader, txm sharedtx.Manager, tokenGenerator memberPorts.TokenGenerator, auditRepo memberPorts.AccessAuditRepository, clock memberPorts.Clock, invitationTTL time.Duration) *Service {
	if invitationTTL <= 0 {
		invitationTTL = defaultOrganizationInvitationTTL
	}
	return &Service{
		repo:           memberRepo,
		roleRepo:       roleRepo,
		accounts:       accountRepo,
		orgReader:      orgReaderAdapter{repo: orgRepo},
		tx:             txm,
		tokenGenerator: tokenGenerator,
		audit:          auditRepo,
		clock:          clock,
		invitationTTL:  invitationTTL,
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
	targetRoleCode := parseRole(role, memberDomain.MembershipRoleMember)
	if strings.TrimSpace(string(targetRoleCode)) == "" {
		return nil, memberErrors.InvalidInput("Invalid role")
	}
	targetResolved, err := s.ResolveRoleForPermissions(ctx, orgID, string(targetRoleCode))
	if err != nil {
		return nil, memberErrors.Internal("resolve role", err)
	}
	if targetResolved == "" {
		return nil, memberErrors.InvalidInput("Invalid role")
	}
	actorResolved, err := s.ResolveRoleForPermissions(ctx, orgID, string(actorMembership.Role()))
	if err != nil || actorResolved == "" {
		return nil, fault.Forbidden("Membership role assignment is not allowed")
	}
	if !accesspolicy.CanAssignOrganizationRole(actorResolved, targetResolved) {
		return nil, fault.Forbidden("Membership role assignment is not allowed")
	}

	var out *memberDomain.MemberView
	err = s.withinTransaction(ctx, func(ctx context.Context) error {
		existing, err := s.repo.GetMemberByAccount(ctx, orgID, targetAccountID)
		if err != nil {
			return memberErrors.Internal("get member by account", err)
		}

		now := s.clock.Now()
		if existing == nil {
			created, err := memberDomain.NewMembership(memberDomain.NewMembershipParams{
				OrganizationID: orgID,
				AccountID:      targetAccountID,
				Role:           targetRoleCode,
				Now:            now,
			})
			if err != nil {
				return memberErrors.InvalidInput("Invalid membership")
			}
			if err := s.repo.AddMember(ctx, orgID, created); err != nil {
				return err
			}
			out, err = s.getMemberViewByAccount(ctx, orgID, targetAccountID)
			if err != nil {
				return err
			}
			return s.appendAudit(ctx, accessAuditParams{
				organizationID: orgID.UUID(),
				actorAccountID: actorAccountID,
				action:         "organization.member.add",
				targetType:     "membership",
				targetID:       &out.MembershipID,
				previousState:  map[string]any{},
				nextState:      membershipViewState(*out),
			})
		}

		existingResolved, err := s.ResolveRoleForPermissions(ctx, orgID, string(existing.Role()))
		if err != nil || existingResolved == "" {
			return fault.Forbidden("Membership change is not allowed for the selected member")
		}
		if !accesspolicy.CanManageOrganizationRole(actorResolved, existingResolved) {
			return fault.Forbidden("Membership change is not allowed for the selected member")
		}
		if existing.IsActive() && !existing.IsRemoved() {
			return memberErrors.MemberAlreadyExists()
		}
		previousState := membershipState(existing)
		if err := existing.ChangeRole(targetRoleCode, now); err != nil {
			return memberErrors.InvalidInput("Invalid role")
		}
		if err := existing.Activate(now); err != nil {
			return memberErrors.InvalidInput("Invalid membership state")
		}
		if err := s.repo.SaveMember(ctx, orgID, existing); err != nil {
			return memberErrors.Internal("save member", err)
		}
		out, err = s.getMemberViewByAccount(ctx, orgID, targetAccountID)
		if err != nil {
			return err
		}
		return s.appendAudit(ctx, accessAuditParams{
			organizationID: orgID.UUID(),
			actorAccountID: actorAccountID,
			action:         "organization.member.add",
			targetType:     "membership",
			targetID:       &out.MembershipID,
			previousState:  previousState,
			nextState:      membershipViewState(*out),
		})
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) UpdateMember(ctx context.Context, cmd UpdateMemberCmd) (*memberDomain.MemberView, error) {
	actorMembership, err := s.requireManageableMembership(ctx, cmd.OrganizationID, cmd.ActorAccountID)
	if err != nil {
		return nil, err
	}

	var out *memberDomain.MemberView
	err = s.withinTransaction(ctx, func(ctx context.Context) error {
		target, err := s.repo.GetMemberByID(ctx, cmd.OrganizationID, cmd.MembershipID)
		if err != nil {
			return memberErrors.Internal("get member by id", err)
		}
		if target == nil {
			return fault.NotFound("Membership not found")
		}
		actorResolved, err := s.ResolveRoleForPermissions(ctx, cmd.OrganizationID, string(actorMembership.Role()))
		if err != nil || actorResolved == "" {
			return fault.Forbidden("Membership change is not allowed for the selected member")
		}
		targetResolved, err := s.ResolveRoleForPermissions(ctx, cmd.OrganizationID, string(target.Role()))
		if err != nil || targetResolved == "" {
			return fault.Forbidden("Membership change is not allowed for the selected member")
		}
		if !accesspolicy.CanManageOrganizationRole(actorResolved, targetResolved) {
			return fault.Forbidden("Membership change is not allowed for the selected member")
		}

		now := s.clock.Now()
		nextRole := target.Role()
		nextResolved := targetResolved
		if cmd.Role != nil {
			nextRole = parseRole(*cmd.Role, target.Role())
			if strings.TrimSpace(string(nextRole)) == "" {
				return memberErrors.InvalidInput("Invalid role")
			}
			var err error
			nextResolved, err = s.ResolveRoleForPermissions(ctx, cmd.OrganizationID, string(nextRole))
			if err != nil {
				return memberErrors.Internal("resolve role", err)
			}
			if nextResolved == "" {
				return memberErrors.InvalidInput("Invalid role")
			}
			if !accesspolicy.CanAssignOrganizationRole(actorResolved, nextResolved) {
				return fault.Forbidden("Membership role assignment is not allowed")
			}
		}

		nextActive := target.IsActive()
		if cmd.IsActive != nil {
			nextActive = *cmd.IsActive
		}
		if targetResolved == memberDomain.MembershipRoleOwner && (!nextActive || nextResolved != memberDomain.MembershipRoleOwner) {
			if err := s.ensureAnotherActiveOwner(ctx, cmd.OrganizationID); err != nil {
				return err
			}
		}

		previousState := membershipState(target)
		if nextRole != target.Role() {
			if err := target.ChangeRole(nextRole, now); err != nil {
				return memberErrors.InvalidInput("Invalid role")
			}
		}
		if cmd.IsActive != nil {
			if *cmd.IsActive {
				if err := target.Activate(now); err != nil {
					return memberErrors.InvalidInput("Invalid membership state")
				}
			} else {
				if err := target.Suspend(now); err != nil {
					return memberErrors.InvalidInput("Invalid membership state")
				}
			}
		}

		if err := s.repo.SaveMember(ctx, cmd.OrganizationID, target); err != nil {
			return memberErrors.Internal("save member", err)
		}
		out, err = s.getMemberViewByID(ctx, cmd.OrganizationID, cmd.MembershipID)
		if err != nil {
			return err
		}
		return s.appendAudit(ctx, accessAuditParams{
			organizationID: cmd.OrganizationID.UUID(),
			actorAccountID: cmd.ActorAccountID,
			action:         "organization.member.update",
			targetType:     "membership",
			targetID:       &out.MembershipID,
			previousState:  previousState,
			nextState:      membershipViewState(*out),
		})
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) RemoveMember(ctx context.Context, cmd RemoveMemberCmd) error {
	actorMembership, err := s.requireManageableMembership(ctx, cmd.OrganizationID, cmd.ActorAccountID)
	if err != nil {
		return err
	}
	return s.withinTransaction(ctx, func(ctx context.Context) error {
		target, err := s.repo.GetMemberByID(ctx, cmd.OrganizationID, cmd.MembershipID)
		if err != nil {
			return memberErrors.Internal("get member by id", err)
		}
		if target == nil {
			return fault.NotFound("Membership not found")
		}
		actorResolved, err := s.ResolveRoleForPermissions(ctx, cmd.OrganizationID, string(actorMembership.Role()))
		if err != nil || actorResolved == "" {
			return fault.Forbidden("Membership removal is not allowed for the selected member")
		}
		targetResolved, err := s.ResolveRoleForPermissions(ctx, cmd.OrganizationID, string(target.Role()))
		if err != nil || targetResolved == "" {
			return fault.Forbidden("Membership removal is not allowed for the selected member")
		}
		if !accesspolicy.CanManageOrganizationRole(actorResolved, targetResolved) {
			return fault.Forbidden("Membership removal is not allowed for the selected member")
		}
		if targetResolved == memberDomain.MembershipRoleOwner && target.IsActive() {
			if err := s.ensureAnotherActiveOwner(ctx, cmd.OrganizationID); err != nil {
				return err
			}
		}
		previousState := membershipState(target)
		if err := target.Remove(s.clock.Now()); err != nil {
			return memberErrors.InvalidInput("Invalid membership state")
		}
		if err := s.repo.SaveMember(ctx, cmd.OrganizationID, target); err != nil {
			return memberErrors.Internal("save member", err)
		}
		nextState := membershipState(target)
		targetID := cmd.MembershipID
		return s.appendAudit(ctx, accessAuditParams{
			organizationID: cmd.OrganizationID.UUID(),
			actorAccountID: cmd.ActorAccountID,
			action:         "organization.member.remove",
			targetType:     "membership",
			targetID:       &targetID,
			previousState:  previousState,
			nextState:      nextState,
		})
	})
}

func (s *Service) ListMembers(ctx context.Context, actorAccountID uuid.UUID, orgID orgDomain.OrganizationID) ([]memberDomain.MemberView, error) {
	if _, err := s.requireReadableMembership(ctx, orgID, actorAccountID); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(ctx, orgID)
}

func (s *Service) CreateInvitation(ctx context.Context, cmd CreateInvitationCmd) (*CreateInvitationResult, error) {
	actorMembership, err := s.requireManageableMembership(ctx, cmd.OrganizationID, cmd.ActorAccountID)
	if err != nil {
		return nil, err
	}
	email, err := accdomain.NewEmail(cmd.Email)
	if err != nil {
		return nil, memberErrors.InvalidInput("Invalid email")
	}
	roleCode := parseRole(cmd.Role, memberDomain.MembershipRoleMember)
	if strings.TrimSpace(string(roleCode)) == "" {
		return nil, memberErrors.InvalidInput("Invalid role")
	}
	roleResolved, err := s.ResolveRoleForPermissions(ctx, cmd.OrganizationID, string(roleCode))
	if err != nil {
		return nil, memberErrors.Internal("resolve role", err)
	}
	if roleResolved == "" {
		return nil, memberErrors.InvalidInput("Invalid role")
	}
	actorResolved, err := s.ResolveRoleForPermissions(ctx, cmd.OrganizationID, string(actorMembership.Role()))
	if err != nil || actorResolved == "" {
		return nil, fault.Forbidden("Membership role assignment is not allowed")
	}
	if !accesspolicy.CanAssignOrganizationRole(actorResolved, roleResolved) {
		return nil, fault.Forbidden("Membership role assignment is not allowed")
	}

	var result *CreateInvitationResult
	err = s.withinTransaction(ctx, func(ctx context.Context) error {
		now := s.clock.Now()
		if err := s.repo.RevokeExpiredPendingInvitations(ctx, cmd.OrganizationID, email, cmd.ActorAccountID, now); err != nil {
			return memberErrors.Internal("revoke expired invitations", err)
		}

		existingAccount, err := s.accounts.GetByEmail(ctx, email)
		if err != nil {
			return memberErrors.Internal("get account by email", err)
		}
		if existingAccount != nil {
			existingMembership, err := s.repo.GetMemberByAccount(ctx, cmd.OrganizationID, existingAccount.ID())
			if err != nil {
				return memberErrors.Internal("get member by account", err)
			}
			if existingMembership != nil && existingMembership.IsActive() && !existingMembership.IsRemoved() {
				return memberErrors.MemberAlreadyExists()
			}
		}

		rawToken, err := s.tokenGenerator.Generate()
		if err != nil {
			return memberErrors.Internal("generate invitation token", err)
		}
		invitation, err := memberDomain.NewOrganizationInvitation(memberDomain.NewOrganizationInvitationParams{
			OrganizationID:   cmd.OrganizationID,
			Email:            email,
			Role:             roleCode,
			TokenHash:        s.tokenGenerator.Hash(rawToken),
			InviterAccountID: cmd.ActorAccountID,
			ExpiresAt:        now.Add(s.invitationTTL),
			Now:              now,
		})
		if err != nil {
			return memberErrors.InvalidInput("Invalid organization invitation")
		}
		if err := s.repo.CreateInvitation(ctx, invitation); err != nil {
			return err
		}
		view := invitation.ToView(now)
		if err := s.appendAudit(ctx, accessAuditParams{
			organizationID: cmd.OrganizationID.UUID(),
			actorAccountID: cmd.ActorAccountID,
			action:         "organization.invitation.create",
			targetType:     "organization_invitation",
			targetID:       uuidPtr(view.ID),
			previousState:  map[string]any{},
			nextState:      invitationViewState(view),
		}); err != nil {
			return err
		}
		result = &CreateInvitationResult{
			Invitation: view,
			Token:      rawToken,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) ListInvitations(ctx context.Context, actorAccountID uuid.UUID, orgID orgDomain.OrganizationID) ([]memberDomain.InvitationView, error) {
	if _, err := s.requireManageableMembership(ctx, orgID, actorAccountID); err != nil {
		return nil, err
	}
	invitations, err := s.repo.ListInvitations(ctx, orgID)
	if err != nil {
		return nil, memberErrors.Internal("list invitations", err)
	}
	now := s.clock.Now()
	out := make([]memberDomain.InvitationView, 0, len(invitations))
	for _, invitation := range invitations {
		out = append(out, invitation.ToView(now))
	}
	return out, nil
}

func (s *Service) AcceptInvitation(ctx context.Context, cmd AcceptInvitationCmd) (*AcceptInvitationResult, error) {
	token := strings.TrimSpace(cmd.Token)
	if token == "" {
		return nil, memberErrors.InvalidInput("Invitation token is required")
	}
	accountID, err := accdomain.AccountIDFromUUID(cmd.ActorAccountID)
	if err != nil || accountID.IsZero() {
		return nil, fault.Unauthorized("Authentication required")
	}

	var result *AcceptInvitationResult
	err = s.withinTransaction(ctx, func(ctx context.Context) error {
		now := s.clock.Now()
		account, err := s.accounts.GetByID(ctx, accountID)
		if err != nil {
			return memberErrors.Internal("get actor account", err)
		}
		if account == nil {
			return fault.Unauthorized("Authentication required")
		}

		invitation, err := s.repo.GetInvitationByTokenHash(ctx, s.tokenGenerator.Hash(token))
		if err != nil {
			return memberErrors.Internal("get invitation by token", err)
		}
		if invitation == nil {
			return memberErrors.InvitationNotFound()
		}
		status := invitation.Status(now)
		if status == memberDomain.InvitationStatusExpired {
			return memberErrors.InvitationExpired()
		}
		if status != memberDomain.InvitationStatusPending {
			return fault.Conflict("Organization invitation is already processed")
		}
		if account.Email() != invitation.Email() {
			return memberErrors.InvitationEmailMismatch()
		}

		existingMembership, err := s.repo.GetMemberByAccount(ctx, invitation.OrganizationID(), account.ID())
		if err != nil {
			return memberErrors.Internal("get member by account", err)
		}
		var memberView *memberDomain.MemberView
		if existingMembership == nil {
			member, err := memberDomain.NewMembership(memberDomain.NewMembershipParams{
				OrganizationID: invitation.OrganizationID(),
				AccountID:      account.ID(),
				Role:           invitation.Role(),
				Now:            now,
			})
			if err != nil {
				return memberErrors.InvalidInput("Invalid membership")
			}
			if err := s.repo.AddMember(ctx, invitation.OrganizationID(), member); err != nil {
				return err
			}
			memberView, err = s.getMemberViewByAccount(ctx, invitation.OrganizationID(), account.ID())
			if err != nil {
				return err
			}
			if err := s.appendAudit(ctx, accessAuditParams{
				organizationID: invitation.OrganizationID().UUID(),
				actorAccountID: cmd.ActorAccountID,
				action:         "organization.member.add",
				targetType:     "membership",
				targetID:       &memberView.MembershipID,
				previousState:  map[string]any{},
				nextState:      membershipViewState(*memberView),
			}); err != nil {
				return err
			}
		} else {
			if existingMembership.IsActive() && !existingMembership.IsRemoved() {
				return memberErrors.MemberAlreadyExists()
			}
			previousState := membershipState(existingMembership)
			if err := existingMembership.ChangeRole(invitation.Role(), now); err != nil {
				return memberErrors.InvalidInput("Invalid role")
			}
			if err := existingMembership.Activate(now); err != nil {
				return memberErrors.InvalidInput("Invalid membership state")
			}
			if err := s.repo.SaveMember(ctx, invitation.OrganizationID(), existingMembership); err != nil {
				return memberErrors.Internal("save member", err)
			}
			memberView, err = s.getMemberViewByAccount(ctx, invitation.OrganizationID(), account.ID())
			if err != nil {
				return err
			}
			if err := s.appendAudit(ctx, accessAuditParams{
				organizationID: invitation.OrganizationID().UUID(),
				actorAccountID: cmd.ActorAccountID,
				action:         "organization.member.add",
				targetType:     "membership",
				targetID:       &memberView.MembershipID,
				previousState:  previousState,
				nextState:      membershipViewState(*memberView),
			}); err != nil {
				return err
			}
		}

		previousInvitationState := invitationViewState(invitation.ToView(now))
		if err := invitation.Accept(cmd.ActorAccountID, now); err != nil {
			return memberErrors.InvalidInput("Invalid organization invitation")
		}
		if err := s.repo.SaveInvitation(ctx, invitation); err != nil {
			return memberErrors.Internal("save invitation", err)
		}
		invitationView := invitation.ToView(now)
		if err := s.appendAudit(ctx, accessAuditParams{
			organizationID: invitation.OrganizationID().UUID(),
			actorAccountID: cmd.ActorAccountID,
			action:         "organization.invitation.accept",
			targetType:     "organization_invitation",
			targetID:       uuidPtr(invitationView.ID),
			previousState:  previousInvitationState,
			nextState:      invitationViewState(invitationView),
		}); err != nil {
			return err
		}

		result = &AcceptInvitationResult{
			Invitation: invitationView,
			Member:     *memberView,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) CreateAccessRequest(ctx context.Context, cmd CreateAccessRequestCmd) (*memberDomain.AccessRequestView, error) {
	if cmd.OrganizationID.IsZero() {
		return nil, memberErrors.InvalidInput("Invalid organization_id")
	}
	requesterID, err := accdomain.AccountIDFromUUID(cmd.ActorAccountID)
	if err != nil || requesterID.IsZero() {
		return nil, fault.Unauthorized("Authentication required")
	}
	exists, err := s.orgReader.Exists(ctx, cmd.OrganizationID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, memberErrors.OrganizationNotFound()
	}

	roleCode := parseRole(cmd.Role, memberDomain.MembershipRoleMember)
	if strings.TrimSpace(string(roleCode)) == "" {
		return nil, memberErrors.InvalidInput("Invalid role")
	}
	roleResolved, err := s.ResolveRoleForPermissions(ctx, cmd.OrganizationID, string(roleCode))
	if err != nil {
		return nil, memberErrors.Internal("resolve role", err)
	}
	if roleResolved == "" {
		return nil, memberErrors.InvalidInput("Invalid role")
	}

	var out *memberDomain.AccessRequestView
	err = s.withinTransaction(ctx, func(ctx context.Context) error {
		member, err := s.repo.GetMemberByAccount(ctx, cmd.OrganizationID, requesterID)
		if err != nil {
			return memberErrors.Internal("get requester membership", err)
		}
		if member != nil && member.IsActive() && !member.IsRemoved() {
			return memberErrors.MemberAlreadyExists()
		}

		req, err := memberDomain.NewOrganizationAccessRequest(memberDomain.NewOrganizationAccessRequestParams{
			OrganizationID:   cmd.OrganizationID,
			RequesterAccount: cmd.ActorAccountID,
			RequestedRole:    roleCode,
			Message:          cmd.Message,
			Now:              s.clock.Now(),
		})
		if err != nil {
			return memberErrors.InvalidInput("Invalid organization access request")
		}
		if err := s.repo.CreateAccessRequest(ctx, req); err != nil {
			return err
		}
		view := req.ToView()
		out = &view
		return s.appendAudit(ctx, accessAuditParams{
			organizationID: cmd.OrganizationID.UUID(),
			actorAccountID: cmd.ActorAccountID,
			action:         "organization.access_request.create",
			targetType:     "organization_access_request",
			targetID:       uuidPtr(view.ID),
			previousState:  map[string]any{},
			nextState:      accessRequestViewState(view),
		})
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) ListAccessRequests(ctx context.Context, actorAccountID uuid.UUID, orgID orgDomain.OrganizationID) ([]memberDomain.AccessRequestView, error) {
	if _, err := s.requireManageableMembership(ctx, orgID, actorAccountID); err != nil {
		return nil, err
	}
	requests, err := s.repo.ListAccessRequests(ctx, orgID)
	if err != nil {
		return nil, memberErrors.Internal("list access requests", err)
	}
	out := make([]memberDomain.AccessRequestView, 0, len(requests))
	for _, req := range requests {
		out = append(out, req.ToView())
	}
	return out, nil
}

func (s *Service) ReviewAccessRequest(ctx context.Context, cmd ReviewAccessRequestCmd) (*memberDomain.AccessRequestView, *memberDomain.MemberView, error) {
	actorMembership, err := s.requireManageableMembership(ctx, cmd.OrganizationID, cmd.ActorAccountID)
	if err != nil {
		return nil, nil, err
	}
	decision := AccessRequestReviewDecision(strings.ToLower(strings.TrimSpace(cmd.Decision)))
	if decision != AccessRequestDecisionApprove && decision != AccessRequestDecisionReject {
		return nil, nil, memberErrors.InvalidInput("Invalid access request decision")
	}

	var requestView *memberDomain.AccessRequestView
	var memberView *memberDomain.MemberView
	err = s.withinTransaction(ctx, func(ctx context.Context) error {
		req, err := s.repo.GetAccessRequestByID(ctx, cmd.OrganizationID, cmd.RequestID)
		if err != nil {
			return memberErrors.Internal("get access request", err)
		}
		if req == nil {
			return memberErrors.AccessRequestNotFound()
		}
		if req.Status() != memberDomain.AccessRequestStatusPending {
			return memberErrors.AccessRequestAlreadyReviewed()
		}

		previousState := accessRequestViewState(req.ToView())
		now := s.clock.Now()

		if decision == AccessRequestDecisionApprove {
			actorResolved, err := s.ResolveRoleForPermissions(ctx, cmd.OrganizationID, string(actorMembership.Role()))
			if err != nil || actorResolved == "" {
				return fault.Forbidden("Membership role assignment is not allowed")
			}
			requestedResolved, err := s.ResolveRoleForPermissions(ctx, cmd.OrganizationID, string(req.RequestedRole()))
			if err != nil || requestedResolved == "" {
				return fault.Forbidden("Membership role assignment is not allowed")
			}
			if !accesspolicy.CanAssignOrganizationRole(actorResolved, requestedResolved) {
				return fault.Forbidden("Membership role assignment is not allowed")
			}
			targetAccountID, err := accdomain.AccountIDFromUUID(req.RequesterAccountID())
			if err != nil || targetAccountID.IsZero() {
				return memberErrors.InvalidInput("Invalid requester account")
			}
			existing, err := s.repo.GetMemberByAccount(ctx, cmd.OrganizationID, targetAccountID)
			if err != nil {
				return memberErrors.Internal("get requester membership", err)
			}
			if existing == nil {
				member, err := memberDomain.NewMembership(memberDomain.NewMembershipParams{
					OrganizationID: cmd.OrganizationID,
					AccountID:      targetAccountID,
					Role:           req.RequestedRole(),
					Now:            now,
				})
				if err != nil {
					return memberErrors.InvalidInput("Invalid membership")
				}
				if err := s.repo.AddMember(ctx, cmd.OrganizationID, member); err != nil {
					return err
				}
				memberView, err = s.getMemberViewByAccount(ctx, cmd.OrganizationID, targetAccountID)
				if err != nil {
					return err
				}
			} else {
				if existing.IsActive() && !existing.IsRemoved() {
					return memberErrors.MemberAlreadyExists()
				}
				if err := existing.ChangeRole(req.RequestedRole(), now); err != nil {
					return memberErrors.InvalidInput("Invalid role")
				}
				if err := existing.Activate(now); err != nil {
					return memberErrors.InvalidInput("Invalid membership state")
				}
				if err := s.repo.SaveMember(ctx, cmd.OrganizationID, existing); err != nil {
					return memberErrors.Internal("save member", err)
				}
				memberView, err = s.getMemberViewByAccount(ctx, cmd.OrganizationID, targetAccountID)
				if err != nil {
					return err
				}
			}
			if err := req.Approve(cmd.ActorAccountID, cmd.ReviewNote, now); err != nil {
				return memberErrors.InvalidInput("Invalid organization access request")
			}
		} else {
			if err := req.Reject(cmd.ActorAccountID, cmd.ReviewNote, now); err != nil {
				return memberErrors.InvalidInput("Invalid organization access request")
			}
		}

		if err := s.repo.SaveAccessRequest(ctx, req); err != nil {
			return memberErrors.Internal("save access request", err)
		}

		view := req.ToView()
		requestView = &view

		if err := s.appendAudit(ctx, accessAuditParams{
			organizationID: cmd.OrganizationID.UUID(),
			actorAccountID: cmd.ActorAccountID,
			action:         "organization.access_request.review",
			targetType:     "organization_access_request",
			targetID:       uuidPtr(view.ID),
			previousState:  previousState,
			nextState:      accessRequestViewState(view),
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return requestView, memberView, nil
}

func (s *Service) requireReadableMembership(ctx context.Context, orgID orgDomain.OrganizationID, actorAccountID uuid.UUID) (*memberDomain.Membership, error) {
	if orgID.IsZero() {
		return nil, memberErrors.InvalidInput("Invalid organization_id")
	}
	actorID, err := accdomain.AccountIDFromUUID(actorAccountID)
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
	resolved, err := s.ResolveRoleForPermissions(ctx, orgID, string(membership.Role()))
	if err != nil || resolved == "" {
		return nil, fault.Forbidden("Organization access denied")
	}
	if !accesspolicy.HasOrganizationPermission(resolved, accesspolicy.PermissionOrganizationRead) {
		return nil, fault.Forbidden("Organization access denied")
	}
	return membership, nil
}

func (s *Service) requireManageableMembership(ctx context.Context, orgID orgDomain.OrganizationID, actorAccountID uuid.UUID) (*memberDomain.Membership, error) {
	membership, err := s.requireReadableMembership(ctx, orgID, actorAccountID)
	if err != nil {
		return nil, err
	}
	resolved, err := s.ResolveRoleForPermissions(ctx, orgID, string(membership.Role()))
	if err != nil || resolved == "" {
		return nil, fault.Forbidden("Only organization owners or admins can manage members")
	}
	if !accesspolicy.HasOrganizationPermission(resolved, accesspolicy.PermissionOrganizationManageMembers) {
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

func (s *Service) getMemberViewByAccount(ctx context.Context, orgID orgDomain.OrganizationID, accountID accdomain.AccountID) (*memberDomain.MemberView, error) {
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

type accessAuditParams struct {
	organizationID uuid.UUID
	actorAccountID uuid.UUID
	action         string
	targetType     string
	targetID       *uuid.UUID
	previousState  map[string]any
	nextState      map[string]any
}

func (s *Service) appendAudit(ctx context.Context, params accessAuditParams) error {
	if s.audit == nil {
		return nil
	}
	var requestID *string
	if value, ok := requestctx.RequestID(ctx); ok {
		requestID = &value
	}
	event := memberDomain.AccessAuditEvent{
		OrganizationID:   params.organizationID,
		ActorSubjectType: "account",
		ActorSubjectID:   uuidPtr(params.actorAccountID),
		ActorAccountID:   uuidPtr(params.actorAccountID),
		Action:           params.action,
		TargetType:       params.targetType,
		TargetID:         params.targetID,
		RequestID:        requestID,
		PreviousState:    params.previousState,
		NextState:        params.nextState,
		CreatedAt:        s.clock.Now(),
	}
	return s.audit.Append(ctx, event)
}

func (s *Service) withinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if s.tx == nil {
		return fn(ctx)
	}
	return s.tx.WithinTransaction(ctx, fn)
}

func membershipState(membership *memberDomain.Membership) map[string]any {
	if membership == nil {
		return map[string]any{}
	}
	return map[string]any{
		"membershipId": membership.ID(),
		"accountId":    membership.AccountID().UUID(),
		"role":         string(membership.Role()),
		"isActive":     membership.IsActive(),
		"deletedAt":    membership.DeletedAt(),
	}
}

func membershipViewState(member memberDomain.MemberView) map[string]any {
	return map[string]any{
		"membershipId": member.MembershipID,
		"accountId":    member.AccountID,
		"role":         member.Role,
		"isActive":     member.IsActive,
		"deletedAt":    member.DeletedAt,
	}
}

func invitationViewState(invitation memberDomain.InvitationView) map[string]any {
	return map[string]any{
		"invitationId": invitation.ID,
		"email":        invitation.Email,
		"role":         invitation.Role,
		"status":       invitation.Status,
		"expiresAt":    invitation.ExpiresAt,
		"acceptedAt":   invitation.AcceptedAt,
	}
}

func accessRequestViewState(request memberDomain.AccessRequestView) map[string]any {
	return map[string]any{
		"requestId":          request.ID,
		"requesterAccountId": request.RequesterAccount,
		"requestedRole":      request.RequestedRole,
		"status":             request.Status,
		"reviewerAccountId":  request.ReviewerAccount,
		"reviewedAt":         request.ReviewedAt,
	}
}

func parseAccountID(raw string) (accdomain.AccountID, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || parsed == uuid.Nil {
		return accdomain.AccountID{}, memberErrors.ErrValidation
	}
	return accdomain.AccountIDFromUUID(parsed)
}

func (s *Service) ListOrganizationRoles(ctx context.Context, q ListOrganizationRolesQuery) ([]*memberDomain.OrganizationRole, error) {
	if _, err := s.requireManageableMembership(ctx, q.OrganizationID, q.ActorAccountID); err != nil {
		return nil, err
	}
	return s.roleRepo.List(ctx, q.OrganizationID, q.IncludeDeleted)
}

func (s *Service) CreateOrganizationRole(ctx context.Context, cmd CreateOrganizationRoleCmd) (*memberDomain.OrganizationRole, error) {
	if _, err := s.requireManageableMembership(ctx, cmd.OrganizationID, cmd.ActorAccountID); err != nil {
		return nil, err
	}
	role, err := memberDomain.NewOrganizationRole(memberDomain.NewOrganizationRoleParams{
		ID:             uuid.New(),
		OrganizationID: cmd.OrganizationID,
		Code:           cmd.Code,
		Name:           cmd.Name,
		Description:    cmd.Description,
		BaseRole:       cmd.BaseRole,
		Now:            s.clock.Now(),
	})
	if err != nil {
		return nil, memberErrors.FromDomain(err)
	}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, memberErrors.FromDomain(err)
	}
	return role, nil
}

func (s *Service) UpdateOrganizationRole(ctx context.Context, cmd UpdateOrganizationRoleCmd) (*memberDomain.OrganizationRole, error) {
	if _, err := s.requireManageableMembership(ctx, cmd.OrganizationID, cmd.ActorAccountID); err != nil {
		return nil, err
	}
	role, err := s.roleRepo.GetByID(ctx, cmd.OrganizationID, cmd.RoleID)
	if err != nil {
		return nil, err
	}
	if role == nil || role.IsDeleted() {
		return nil, fault.NotFound("Organization role not found", fault.Code("ORGANIZATION_ROLE_NOT_FOUND"))
	}
	patch := memberDomain.OrganizationRolePatch{UpdatedAt: s.clock.Now()}
	if cmd.Name != nil {
		patch.Name = cmd.Name
	}
	if cmd.Description != nil {
		patch.Description = cmd.Description
	}
	if cmd.BaseRole != nil {
		patch.BaseRole = cmd.BaseRole
	}
	if err := role.ApplyPatch(patch); err != nil {
		return nil, memberErrors.FromDomain(err)
	}
	if err := s.roleRepo.Save(ctx, role); err != nil {
		return nil, memberErrors.FromDomain(err)
	}
	return role, nil
}

func (s *Service) DeleteOrganizationRole(ctx context.Context, cmd DeleteOrganizationRoleCmd) error {
	if _, err := s.requireManageableMembership(ctx, cmd.OrganizationID, cmd.ActorAccountID); err != nil {
		return err
	}
	role, err := s.roleRepo.GetByID(ctx, cmd.OrganizationID, cmd.RoleID)
	if err != nil {
		return err
	}
	if role == nil || role.IsDeleted() {
		return fault.NotFound("Organization role not found", fault.Code("ORGANIZATION_ROLE_NOT_FOUND"))
	}
	count, err := s.roleRepo.CountMembersWithRole(ctx, cmd.OrganizationID, role.Code())
	if err != nil {
		return fault.Internal("Count members failed", fault.WithCause(err))
	}
	if count > 0 {
		return fault.Conflict("Role is in use and cannot be deleted. Reassign or remove members first.", fault.Code("ORGANIZATION_ROLE_IN_USE"))
	}
	now := s.clock.Now()
	if err := role.SoftDelete(now); err != nil {
		return memberErrors.FromDomain(err)
	}
	return s.roleRepo.Save(ctx, role)
}

func (s *Service) ResolveRoleForPermissions(ctx context.Context, orgID orgDomain.OrganizationID, roleCode string) (memberDomain.MembershipRole, error) {
	r := memberDomain.MembershipRole(strings.ToLower(strings.TrimSpace(roleCode)))
	if r.IsValid() {
		return r, nil
	}
	custom, err := s.roleRepo.GetByCode(ctx, orgID, roleCode)
	if err != nil {
		return "", err
	}
	if custom == nil || custom.IsDeleted() {
		return "", nil
	}
	return custom.BaseRole(), nil
}

func parseRole(raw string, fallback memberDomain.MembershipRole) memberDomain.MembershipRole {
	parsed := memberDomain.ParseMembershipRole(raw)
	if parsed == "" {
		return fallback
	}
	return parsed
}

func uuidPtr(value uuid.UUID) *uuid.UUID {
	if value == uuid.Nil {
		return nil
	}
	v := value
	return &v
}
