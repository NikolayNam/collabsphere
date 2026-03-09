package http

import (
	"context"
	"errors"
	"net/http"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	groupsapp "github.com/NikolayNam/collabsphere/internal/groups/application"
	"github.com/NikolayNam/collabsphere/internal/groups/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/groups/delivery/http/mapper"
	groupsdomain "github.com/NikolayNam/collabsphere/internal/groups/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/httpbind"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
)

type Handler struct {
	svc *groupsapp.Service
}

func NewHandler(svc *groupsapp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateGroup(ctx context.Context, input *dto.CreateGroupInput) (*dto.GroupResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	group, err := h.svc.CreateGroup(ctx, groupsapp.CreateGroupCmd{
		Name:           input.Body.Name,
		Slug:           input.Body.Slug,
		Description:    input.Body.Description,
		OwnerAccountID: actorID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToGroupResponse(group, http.StatusCreated), nil
}

func (h *Handler) GetGroupByID(ctx context.Context, input *dto.GetGroupByIDInput) (*dto.GroupResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	groupID, err := parseGroupID(input.ID)
	if err != nil {
		return nil, humaerr.From(ctx, groupsapp.ErrValidation)
	}

	group, err := h.svc.GetGroupByID(ctx, groupsapp.GetGroupByIDQuery{ID: groupID, ActorAccountID: actorID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToGroupResponse(group, http.StatusOK), nil
}

func (h *Handler) AddAccountMember(ctx context.Context, input *dto.AddAccountMemberInput) (*dto.GroupAccountMemberResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	groupID, err := parseGroupID(input.GroupID)
	if err != nil {
		return nil, humaerr.From(ctx, groupsapp.ErrValidation)
	}
	accountID, err := parseAccountID(input.Body.AccountID)
	if err != nil {
		return nil, humaerr.From(ctx, groupsapp.ErrValidation)
	}

	if err := h.svc.AddAccountMember(ctx, groupsapp.AddAccountMemberCmd{
		GroupID:        groupID,
		ActorAccountID: actorID,
		AccountID:      accountID,
		Role:           input.Body.Role,
	}); err != nil {
		return nil, humaerr.From(ctx, err)
	}

	members, err := h.svc.ListMembers(ctx, groupsapp.ListMembersQuery{GroupID: groupID, ActorAccountID: actorID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	for _, member := range members.Accounts {
		if member.AccountID == accountID.UUID() {
			return mapper.ToAccountMemberResponse(member), nil
		}
	}
	return nil, humaerr.From(ctx, errors.New("group account member not found after successful add"))
}

func (h *Handler) AddOrganizationMember(ctx context.Context, input *dto.AddOrganizationMemberInput) (*dto.GroupOrganizationMemberResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	groupID, err := parseGroupID(input.GroupID)
	if err != nil {
		return nil, humaerr.From(ctx, groupsapp.ErrValidation)
	}
	organizationID, err := parseOrganizationID(input.Body.OrganizationID)
	if err != nil {
		return nil, humaerr.From(ctx, groupsapp.ErrValidation)
	}

	if err := h.svc.AddOrganizationMember(ctx, groupsapp.AddOrganizationMemberCmd{
		GroupID:        groupID,
		ActorAccountID: actorID,
		OrganizationID: organizationID,
	}); err != nil {
		return nil, humaerr.From(ctx, err)
	}

	members, err := h.svc.ListMembers(ctx, groupsapp.ListMembersQuery{GroupID: groupID, ActorAccountID: actorID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	for _, member := range members.Organizations {
		if member.OrganizationID == organizationID.UUID() {
			return mapper.ToOrganizationMemberResponse(member), nil
		}
	}
	return nil, humaerr.From(ctx, errors.New("group organization member not found after successful add"))
}

func (h *Handler) ListMembers(ctx context.Context, input *dto.ListMembersInput) (*dto.GroupMembersResponse, error) {
	actorID, err := currentActorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	groupID, err := parseGroupID(input.GroupID)
	if err != nil {
		return nil, humaerr.From(ctx, groupsapp.ErrValidation)
	}

	members, err := h.svc.ListMembers(ctx, groupsapp.ListMembersQuery{GroupID: groupID, ActorAccountID: actorID})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToMembersResponse(members), nil
}

func currentActorAccountID(ctx context.Context) (accdomain.AccountID, error) {
	return httpbind.RequireAccountID(ctx, fault.Unauthorized("Authentication required"))
}

func parseGroupID(value string) (groupsdomain.GroupID, error) {
	return httpbind.ParseGroupID(value, groupsapp.ErrValidation)
}

func parseAccountID(value string) (accdomain.AccountID, error) {
	return httpbind.ParseAccountID(value, groupsapp.ErrValidation)
}

func parseOrganizationID(value string) (orgdomain.OrganizationID, error) {
	return httpbind.ParseOrganizationID(value, groupsapp.ErrValidation)
}
