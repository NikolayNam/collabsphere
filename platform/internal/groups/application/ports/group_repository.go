package ports

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
)

type GroupRepository interface {
	Create(ctx context.Context, group *domain.Group) error
	GetByID(ctx context.Context, id domain.GroupID) (*domain.Group, error)
	GetAccountMember(ctx context.Context, groupID domain.GroupID, accountID accdomain.AccountID) (*domain.AccountMember, error)
	AddAccountMember(ctx context.Context, groupID domain.GroupID, member *domain.AccountMember) error
	AddOrganizationMember(ctx context.Context, groupID domain.GroupID, member *domain.OrganizationMember) error
	ListMembers(ctx context.Context, groupID domain.GroupID) (*domain.MembersView, error)
}
