package application

import (
	"context"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/groups/application/add_account_member"
	"github.com/NikolayNam/collabsphere/internal/groups/application/add_organization_member"
	"github.com/NikolayNam/collabsphere/internal/groups/application/create_group"
	create_with_owner "github.com/NikolayNam/collabsphere/internal/groups/application/create_group_with_owner"
	"github.com/NikolayNam/collabsphere/internal/groups/application/errors"
	"github.com/NikolayNam/collabsphere/internal/groups/application/get_group_by_id"
	"github.com/NikolayNam/collabsphere/internal/groups/application/list_members"
	"github.com/NikolayNam/collabsphere/internal/groups/application/ports"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	sharedtx "github.com/NikolayNam/collabsphere/shared/tx"
)

var (
	ErrValidation = errors.ErrValidation
)

type CreateGroupCmd = create_group.Command
type AddAccountMemberCmd = add_account_member.Command
type AddOrganizationMemberCmd = add_organization_member.Command
type GetGroupByIDQuery = get_group_by_id.Query
type ListMembersQuery = list_members.Query

type MyGroupView struct {
	ID               string
	Name             string
	Slug             string
	Description      *string
	IsActive         bool
	CreatedAt        string
	MembershipSource string
	MembershipRole   *string
}

type Service struct {
	create                *create_group.Handler
	repo                  ports.GroupRepository
	getByID               *get_group_by_id.Handler
	addAccountMember      *add_account_member.Handler
	addOrganizationMember *add_organization_member.Handler
	listMembers           *list_members.Handler
}

func New(repo ports.GroupRepository, accounts ports.AccountReader, organizations ports.OrganizationReader, channels ports.ChannelProvisioner, txm sharedtx.Manager, clock ports.Clock) *Service {
	creator := create_with_owner.NewHandler(txm, repo, channels)

	return &Service{
		create:                create_group.NewHandler(creator, clock),
		repo:                  repo,
		getByID:               get_group_by_id.NewHandler(repo),
		addAccountMember:      add_account_member.NewHandler(repo, accounts, clock),
		addOrganizationMember: add_organization_member.NewHandler(repo, organizations, clock),
		listMembers:           list_members.NewHandler(repo),
	}
}

func (s *Service) CreateGroup(ctx context.Context, cmd CreateGroupCmd) (*domain.Group, error) {
	return s.create.Handle(ctx, cmd)
}

func (s *Service) GetGroupByID(ctx context.Context, q GetGroupByIDQuery) (*domain.Group, error) {
	return s.getByID.Handle(ctx, q)
}

func (s *Service) ListMyGroups(ctx context.Context, actorAccountID accdomain.AccountID) ([]MyGroupView, error) {
	items, err := s.repo.ListByAccount(ctx, actorAccountID)
	if err != nil {
		return nil, fault.Internal("List groups failed", fault.WithCause(err))
	}

	out := make([]MyGroupView, 0, len(items))
	for _, item := range items {
		out = append(out, MyGroupView{
			ID:               item.ID.String(),
			Name:             item.Name,
			Slug:             item.Slug,
			Description:      item.Description,
			IsActive:         item.IsActive,
			CreatedAt:        item.CreatedAt.Format(time.RFC3339Nano),
			MembershipSource: item.MembershipSource,
			MembershipRole:   item.MembershipRole,
		})
	}
	return out, nil
}

func (s *Service) AddAccountMember(ctx context.Context, cmd AddAccountMemberCmd) error {
	return s.addAccountMember.Handle(ctx, cmd)
}

func (s *Service) AddOrganizationMember(ctx context.Context, cmd AddOrganizationMemberCmd) error {
	return s.addOrganizationMember.Handle(ctx, cmd)
}

func (s *Service) ListMembers(ctx context.Context, q ListMembersQuery) (*domain.MembersView, error) {
	return s.listMembers.Handle(ctx, q)
}
