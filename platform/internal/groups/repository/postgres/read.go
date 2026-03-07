package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
	"github.com/NikolayNam/collabsphere/internal/groups/repository/postgres/dbmodel"
	"gorm.io/gorm"
)

func (r *GroupRepo) GetByID(ctx context.Context, id domain.GroupID) (*domain.Group, error) {
	var model dbmodel.Group
	if err := r.dbFrom(ctx).WithContext(ctx).Take(&model, "id = ?", id.UUID()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return domain.RehydrateGroup(domain.RehydrateGroupParams{
		ID:          id,
		Name:        model.Name,
		Slug:        model.Slug,
		Description: model.Description,
		IsActive:    model.IsActive,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	})
}

func (r *GroupRepo) ListMembers(ctx context.Context, groupID domain.GroupID) (*domain.MembersView, error) {
	if groupID.IsZero() {
		return &domain.MembersView{}, nil
	}

	var accountRows []struct {
		MembershipID uuid.UUID `gorm:"column:membership_id"`
		AccountID    uuid.UUID `gorm:"column:account_id"`
		Email        string    `gorm:"column:email"`
		DisplayName  *string   `gorm:"column:display_name"`
		Role         string    `gorm:"column:role"`
		IsActive     bool      `gorm:"column:is_active"`
		CreatedAt    time.Time `gorm:"column:created_at"`
	}
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.group_account_members AS gam").
		Select("gam.id AS membership_id, gam.account_id, a.email, a.display_name, gam.role, gam.is_active, gam.created_at").
		Joins("JOIN iam.accounts AS a ON a.id = gam.account_id").
		Where("gam.group_id = ?", groupID.UUID()).
		Order("gam.created_at ASC").
		Scan(&accountRows).Error; err != nil {
		return nil, err
	}

	var organizationRows []struct {
		MembershipID   uuid.UUID `gorm:"column:membership_id"`
		OrganizationID uuid.UUID `gorm:"column:organization_id"`
		Name           string    `gorm:"column:name"`
		Slug           string    `gorm:"column:slug"`
		IsActive       bool      `gorm:"column:is_active"`
		CreatedAt      time.Time `gorm:"column:created_at"`
	}
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.group_organization_members AS gom").
		Select("gom.id AS membership_id, gom.organization_id, o.name, o.slug, gom.is_active, gom.created_at").
		Joins("JOIN org.organizations AS o ON o.id = gom.organization_id").
		Where("gom.group_id = ?", groupID.UUID()).
		Order("gom.created_at ASC").
		Scan(&organizationRows).Error; err != nil {
		return nil, err
	}

	view := &domain.MembersView{
		Accounts:      make([]domain.AccountMemberView, 0, len(accountRows)),
		Organizations: make([]domain.OrganizationMemberView, 0, len(organizationRows)),
	}
	for _, row := range accountRows {
		view.Accounts = append(view.Accounts, domain.AccountMemberView{
			MembershipID: row.MembershipID,
			AccountID:    row.AccountID,
			Email:        row.Email,
			DisplayName:  row.DisplayName,
			Role:         row.Role,
			IsActive:     row.IsActive,
			CreatedAt:    row.CreatedAt,
		})
	}
	for _, row := range organizationRows {
		view.Organizations = append(view.Organizations, domain.OrganizationMemberView{
			MembershipID:   row.MembershipID,
			OrganizationID: row.OrganizationID,
			Name:           row.Name,
			Slug:           row.Slug,
			IsActive:       row.IsActive,
			CreatedAt:      row.CreatedAt,
		})
	}

	return view, nil
}
