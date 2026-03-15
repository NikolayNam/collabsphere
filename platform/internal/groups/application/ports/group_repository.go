package ports

import (
	"context"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
	"github.com/google/uuid"
)

type GroupMembershipView struct {
	ID               uuid.UUID
	Name             string
	Slug             string
	Description      *string
	IsActive         bool
	CreatedAt        time.Time
	MembershipSource string
	MembershipRole   *string
}

type GroupRepository interface {
	Create(ctx context.Context, group *domain.Group) error
	GetByID(ctx context.Context, id domain.GroupID) (*domain.Group, error)
	ListByAccount(ctx context.Context, accountID accdomain.AccountID) ([]GroupMembershipView, error)
	HasGroupAccessForAccount(ctx context.Context, groupID domain.GroupID, accountID accdomain.AccountID) (bool, error)
	GetAccountMember(ctx context.Context, groupID domain.GroupID, accountID accdomain.AccountID) (*domain.AccountMember, error)
	AddAccountMember(ctx context.Context, groupID domain.GroupID, member *domain.AccountMember) error
	AddOrganizationMember(ctx context.Context, groupID domain.GroupID, member *domain.OrganizationMember) error
	ListMembers(ctx context.Context, groupID domain.GroupID) (*domain.MembersView, error)
}
