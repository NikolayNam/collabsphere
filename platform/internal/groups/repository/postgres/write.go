package postgres

import (
	"context"
	"errors"
	"fmt"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	groupsErrors "github.com/NikolayNam/collabsphere/internal/groups/application/errors"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
	"github.com/NikolayNam/collabsphere/internal/groups/repository/postgres/dbmodel"
	"gorm.io/gorm"
)

func (r *GroupRepo) Create(ctx context.Context, group *domain.Group) error {
	if group == nil {
		return groupsErrors.InvalidInput("Group is required")
	}

	updatedAt := group.CreatedAt()
	if group.UpdatedAt() != nil {
		updatedAt = *group.UpdatedAt()
	}

	model := &dbmodel.Group{
		ID:          group.ID().UUID(),
		Name:        group.Name(),
		Slug:        group.Slug(),
		Description: group.Description(),
		IsActive:    group.IsActive(),
		CreatedAt:   group.CreatedAt(),
		UpdatedAt:   updatedAt,
	}

	if err := r.dbFrom(ctx).WithContext(ctx).Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: %w", groupsErrors.ErrConflict, err)
		}
		return err
	}
	return nil
}

func (r *GroupRepo) AddAccountMember(ctx context.Context, groupID domain.GroupID, member *domain.AccountMember) error {
	if groupID.IsZero() || member == nil {
		return groupsErrors.InvalidInput("Group account member is required")
	}

	updatedAt := member.CreatedAt()
	if member.UpdatedAt() != nil {
		updatedAt = *member.UpdatedAt()
	}

	model := &dbmodel.GroupAccountMember{
		ID:        member.ID(),
		GroupID:   groupID.UUID(),
		AccountID: member.AccountID().UUID(),
		Role:      string(member.Role()),
		IsActive:  member.IsActive(),
		CreatedAt: member.CreatedAt(),
		UpdatedAt: updatedAt,
	}

	if err := r.dbFrom(ctx).WithContext(ctx).Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return groupsErrors.GroupMemberAlreadyExists()
		}
		if isForeignKeyViolation(err) {
			return groupsErrors.InvalidInput("Group or account not found")
		}
		return err
	}
	return nil
}

func (r *GroupRepo) AddOrganizationMember(ctx context.Context, groupID domain.GroupID, member *domain.OrganizationMember) error {
	if groupID.IsZero() || member == nil {
		return groupsErrors.InvalidInput("Group organization member is required")
	}

	updatedAt := member.CreatedAt()
	if member.UpdatedAt() != nil {
		updatedAt = *member.UpdatedAt()
	}

	model := &dbmodel.GroupOrganizationMember{
		ID:             member.ID(),
		GroupID:        groupID.UUID(),
		OrganizationID: member.OrganizationID().UUID(),
		IsActive:       member.IsActive(),
		CreatedAt:      member.CreatedAt(),
		UpdatedAt:      updatedAt,
	}

	if err := r.dbFrom(ctx).WithContext(ctx).Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return groupsErrors.GroupMemberAlreadyExists()
		}
		if isForeignKeyViolation(err) {
			return groupsErrors.InvalidInput("Group or organization not found")
		}
		return err
	}
	return nil
}

func (r *GroupRepo) GetAccountMember(ctx context.Context, groupID domain.GroupID, accountID accdomain.AccountID) (*domain.AccountMember, error) {
	if groupID.IsZero() || accountID.IsZero() {
		return nil, nil
	}

	var model dbmodel.GroupAccountMember
	if err := r.dbFrom(ctx).WithContext(ctx).
		Take(&model, "group_id = ? AND account_id = ?", groupID.UUID(), accountID.UUID()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	role := domain.GroupAccountRole(model.Role)
	if !role.IsValid() {
		return nil, domain.ErrGroupRoleInvalid
	}

	return domain.RehydrateAccountMember(domain.RehydrateAccountMemberParams{
		ID:        model.ID,
		GroupID:   groupID,
		AccountID: accountID,
		Role:      role,
		IsActive:  model.IsActive,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	})
}
