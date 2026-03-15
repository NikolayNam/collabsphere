package http

import (
	"context"
	"net/http"

	"github.com/NikolayNam/collabsphere/internal/memberships/application"
	"github.com/NikolayNam/collabsphere/internal/memberships/delivery/http/dto"
	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/google/uuid"
)

type Handler struct {
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) AddMember(ctx context.Context, input *dto.AddMemberInput) (*dto.MemberResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orgID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	member, err := h.svc.AddMember(ctx, actorID, orgID, input.Body.AccountID, input.Body.Role)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.MemberResponse{Status: http.StatusCreated, Body: toMemberPayload(*member)}, nil
}

func (h *Handler) UpdateMember(ctx context.Context, input *dto.UpdateMemberInput) (*dto.MemberResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orgID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	membershipID, err := parseUUID(input.MembershipID)
	if err != nil {
		return nil, humaerr.From(ctx, fault.Validation("Invalid membership_id"))
	}
	member, err := h.svc.UpdateMember(ctx, application.UpdateMemberCmd{
		OrganizationID: orgID,
		MembershipID:   membershipID,
		ActorAccountID: actorID,
		Role:           input.Body.Role,
		IsActive:       input.Body.IsActive,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.MemberResponse{Status: http.StatusOK, Body: toMemberPayload(*member)}, nil
}

func (h *Handler) RemoveMember(ctx context.Context, input *dto.RemoveMemberInput) (*dto.EmptyResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orgID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	membershipID, err := parseUUID(input.MembershipID)
	if err != nil {
		return nil, humaerr.From(ctx, fault.Validation("Invalid membership_id"))
	}
	if err := h.svc.RemoveMember(ctx, application.RemoveMemberCmd{OrganizationID: orgID, MembershipID: membershipID, ActorAccountID: actorID}); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.EmptyResponse{Status: http.StatusNoContent}, nil
}

func (h *Handler) ListMembers(ctx context.Context, input *dto.ListMembersInput) (*dto.MembersListResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orgID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	members, err := h.svc.ListMembers(ctx, actorID, orgID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.MembersListResponse{Status: http.StatusOK}
	out.Body.Members = make([]dto.MemberPayload, 0, len(members))
	for _, member := range members {
		out.Body.Members = append(out.Body.Members, toMemberPayload(member))
	}
	return out, nil
}

func (h *Handler) CreateInvitation(ctx context.Context, input *dto.CreateInvitationInput) (*dto.InvitationResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orgID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	res, err := h.svc.CreateInvitation(ctx, application.CreateInvitationCmd{
		OrganizationID: orgID,
		ActorAccountID: actorID,
		Email:          input.Body.Email,
		Role:           input.Body.Role,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	body := toInvitationPayload(res.Invitation)
	body.Token = &res.Token
	return &dto.InvitationResponse{Status: http.StatusCreated, Body: body}, nil
}

func (h *Handler) ListInvitations(ctx context.Context, input *dto.ListInvitationsInput) (*dto.InvitationsListResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orgID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	invitations, err := h.svc.ListInvitations(ctx, actorID, orgID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.InvitationsListResponse{Status: http.StatusOK}
	out.Body.Invitations = make([]dto.InvitationPayload, 0, len(invitations))
	for _, invitation := range invitations {
		out.Body.Invitations = append(out.Body.Invitations, toInvitationPayload(invitation))
	}
	return out, nil
}

func (h *Handler) AcceptInvitation(ctx context.Context, input *dto.AcceptInvitationInput) (*dto.AcceptInvitationResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	res, err := h.svc.AcceptInvitation(ctx, application.AcceptInvitationCmd{
		Token:          input.Token,
		ActorAccountID: actorID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.AcceptInvitationResponse{Status: http.StatusOK}
	out.Body.Invitation = toInvitationPayload(res.Invitation)
	out.Body.Member = toMemberPayload(res.Member)
	return out, nil
}

func (h *Handler) CreateAccessRequest(ctx context.Context, input *dto.CreateAccessRequestInput) (*dto.AccessRequestResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orgID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	req, err := h.svc.CreateAccessRequest(ctx, application.CreateAccessRequestCmd{
		OrganizationID: orgID,
		ActorAccountID: actorID,
		Role:           input.Body.Role,
		Message:        input.Body.Message,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.AccessRequestResponse{Status: http.StatusCreated, Body: toAccessRequestPayload(*req)}, nil
}

func (h *Handler) ListAccessRequests(ctx context.Context, input *dto.ListAccessRequestsInput) (*dto.AccessRequestsListResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orgID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	items, err := h.svc.ListAccessRequests(ctx, actorID, orgID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.AccessRequestsListResponse{Status: http.StatusOK}
	out.Body.Requests = make([]dto.AccessRequestPayload, 0, len(items))
	for _, item := range items {
		out.Body.Requests = append(out.Body.Requests, toAccessRequestPayload(item))
	}
	return out, nil
}

func (h *Handler) ApproveAccessRequest(ctx context.Context, input *dto.ReviewAccessRequestInput) (*dto.ReviewAccessRequestResponse, error) {
	return h.reviewAccessRequest(ctx, input, string(application.AccessRequestDecisionApprove))
}

func (h *Handler) RejectAccessRequest(ctx context.Context, input *dto.ReviewAccessRequestInput) (*dto.ReviewAccessRequestResponse, error) {
	return h.reviewAccessRequest(ctx, input, string(application.AccessRequestDecisionReject))
}

func (h *Handler) ListOrganizationRoles(ctx context.Context, input *dto.ListOrganizationRolesInput) (*dto.OrganizationRolesListResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orgID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	roles, err := h.svc.ListOrganizationRoles(ctx, application.ListOrganizationRolesQuery{
		OrganizationID:  orgID,
		ActorAccountID:  actorID,
		IncludeDeleted:  input.IncludeDeleted,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.OrganizationRolesListResponse{Status: http.StatusOK}
	out.Body.Roles = append(out.Body.Roles, systemRolePayloads(orgID)...)
	for _, r := range roles {
		out.Body.Roles = append(out.Body.Roles, toOrganizationRolePayload(r))
	}
	return out, nil
}

func (h *Handler) CreateOrganizationRole(ctx context.Context, input *dto.CreateOrganizationRoleInput) (*dto.OrganizationRoleResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orgID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	role, err := h.svc.CreateOrganizationRole(ctx, application.CreateOrganizationRoleCmd{
		OrganizationID: orgID,
		ActorAccountID: actorID,
		Code:           input.Body.Code,
		Name:           input.Body.Name,
		Description:   input.Body.Description,
		BaseRole:      input.Body.BaseRole,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.OrganizationRoleResponse{Status: http.StatusCreated, Body: toOrganizationRolePayload(role)}, nil
}

func (h *Handler) UpdateOrganizationRole(ctx context.Context, input *dto.UpdateOrganizationRoleInput) (*dto.OrganizationRoleResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orgID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	roleID, err := parseUUID(input.RoleID)
	if err != nil {
		return nil, humaerr.From(ctx, fault.Validation("Invalid role_id"))
	}
	role, err := h.svc.UpdateOrganizationRole(ctx, application.UpdateOrganizationRoleCmd{
		OrganizationID: orgID,
		RoleID:         roleID,
		ActorAccountID: actorID,
		Name:           input.Body.Name,
		Description:   input.Body.Description,
		BaseRole:      input.Body.BaseRole,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.OrganizationRoleResponse{Status: http.StatusOK, Body: toOrganizationRolePayload(role)}, nil
}

func (h *Handler) DeleteOrganizationRole(ctx context.Context, input *dto.DeleteOrganizationRoleInput) (*dto.EmptyResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orgID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	roleID, err := parseUUID(input.RoleID)
	if err != nil {
		return nil, humaerr.From(ctx, fault.Validation("Invalid role_id"))
	}
	if err := h.svc.DeleteOrganizationRole(ctx, application.DeleteOrganizationRoleCmd{
		OrganizationID: orgID,
		RoleID:         roleID,
		ActorAccountID: actorID,
	}); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.EmptyResponse{Status: http.StatusNoContent}, nil
}

func (h *Handler) reviewAccessRequest(ctx context.Context, input *dto.ReviewAccessRequestInput, decision string) (*dto.ReviewAccessRequestResponse, error) {
	actorID, err := principalMembershipActor(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orgID, err := parseOrganizationID(input.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	requestID, err := parseUUID(input.RequestID)
	if err != nil {
		return nil, humaerr.From(ctx, fault.Validation("Invalid request_id"))
	}
	requestView, memberView, err := h.svc.ReviewAccessRequest(ctx, application.ReviewAccessRequestCmd{
		OrganizationID: orgID,
		RequestID:      requestID,
		ActorAccountID: actorID,
		Decision:       decision,
		ReviewNote:     input.Body.ReviewNote,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.ReviewAccessRequestResponse{Status: http.StatusOK}
	out.Body.Request = toAccessRequestPayload(*requestView)
	if memberView != nil {
		payload := toMemberPayload(*memberView)
		out.Body.Member = &payload
	}
	return out, nil
}

func principalMembershipActor(ctx context.Context) (uuid.UUID, error) {
	principal := authmw.PrincipalFromContext(ctx)
	if !principal.IsAccount() {
		return uuid.Nil, fault.Unauthorized("Authentication required")
	}
	return principal.AccountID, nil
}

func parseOrganizationID(raw string) (orgDomain.OrganizationID, error) {
	parsed, err := parseUUID(raw)
	if err != nil {
		return orgDomain.OrganizationID{}, fault.Validation("Invalid organization_id")
	}
	return orgDomain.OrganizationIDFromUUID(parsed)
}

func parseUUID(raw string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil || parsed == uuid.Nil {
		return uuid.Nil, fault.Validation("Invalid identifier")
	}
	return parsed, nil
}

func toMemberPayload(member memberDomain.MemberView) dto.MemberPayload {
	return dto.MemberPayload{
		ID:             member.MembershipID,
		OrganizationID: member.OrganizationID,
		AccountID:      member.AccountID,
		Role:           member.Role,
		IsActive:       member.IsActive,
		CreatedAt:      member.CreatedAt,
		UpdatedAt:      member.UpdatedAt,
		DeletedAt:      member.DeletedAt,
	}
}

func toInvitationPayload(invitation memberDomain.InvitationView) dto.InvitationPayload {
	return dto.InvitationPayload{
		ID:                  invitation.ID,
		OrganizationID:      invitation.OrganizationID,
		Email:               invitation.Email,
		Role:                invitation.Role,
		Status:              invitation.Status,
		InviterAccountID:    invitation.InviterAccountID,
		AcceptedByAccountID: invitation.AcceptedByAccountID,
		AcceptedAt:          invitation.AcceptedAt,
		ExpiresAt:           invitation.ExpiresAt,
		CreatedAt:           invitation.CreatedAt,
		UpdatedAt:           invitation.UpdatedAt,
	}
}

func toAccessRequestPayload(request memberDomain.AccessRequestView) dto.AccessRequestPayload {
	return dto.AccessRequestPayload{
		ID:               request.ID,
		OrganizationID:   request.OrganizationID,
		RequesterAccount: request.RequesterAccount,
		RequestedRole:    request.RequestedRole,
		Message:          request.Message,
		Status:           request.Status,
		ReviewerAccount:  request.ReviewerAccount,
		ReviewNote:       request.ReviewNote,
		ReviewedAt:       request.ReviewedAt,
		CreatedAt:        request.CreatedAt,
		UpdatedAt:        request.UpdatedAt,
	}
}

var systemRoleDisplayNames = map[string]string{
	"owner":   "Owner",
	"admin":   "Administrator",
	"manager": "Manager",
	"member":  "Member",
	"viewer":  "Viewer",
}

func systemRolePayloads(orgID orgDomain.OrganizationID) []dto.OrganizationRolePayload {
	codes := []string{"owner", "admin", "manager", "member", "viewer"}
	out := make([]dto.OrganizationRolePayload, 0, len(codes))
	for _, code := range codes {
		name := code
		if n, ok := systemRoleDisplayNames[code]; ok {
			name = n
		}
		out = append(out, dto.OrganizationRolePayload{
			ID:             uuid.Nil,
			OrganizationID: orgID.UUID(),
			Code:           code,
			Name:           name,
			Description:   "",
			BaseRole:      code,
			IsSystem:      true,
		})
	}
	return out
}

func toOrganizationRolePayload(r *memberDomain.OrganizationRole) dto.OrganizationRolePayload {
	return dto.OrganizationRolePayload{
		ID:             r.ID(),
		OrganizationID: r.OrganizationID().UUID(),
		Code:           r.Code(),
		Name:           r.Name(),
		Description:    r.Description(),
		BaseRole:       string(r.BaseRole()),
		IsSystem:       false,
		CreatedAt:      r.CreatedAt(),
		UpdatedAt:      r.UpdatedAt(),
		DeletedAt:      r.DeletedAt(),
	}
}
