package postgres

import (
	"context"
	"errors"
	"sort"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/groups/application/ports"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
	"github.com/NikolayNam/collabsphere/internal/groups/repository/postgres/dbmodel"
	"github.com/google/uuid"
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

func (r *GroupRepo) ListByAccount(ctx context.Context, accountID accdomain.AccountID) ([]ports.GroupMembershipView, error) {
	if accountID.IsZero() {
		return []ports.GroupMembershipView{}, nil
	}

	type row struct {
		ID               uuid.UUID `gorm:"column:id"`
		Name             string    `gorm:"column:name"`
		Slug             string    `gorm:"column:slug"`
		Description      *string   `gorm:"column:description"`
		IsActive         bool      `gorm:"column:is_active"`
		CreatedAt        time.Time `gorm:"column:created_at"`
		MembershipSource string    `gorm:"column:membership_source"`
		MembershipRole   *string   `gorm:"column:membership_role"`
	}

	const query = `
WITH accessible_groups AS (
	SELECT gam.group_id, 'account'::text AS membership_source, gam.role AS membership_role, 1 AS priority
	FROM iam.group_account_members AS gam
	WHERE gam.account_id = @account_id
	  AND gam.deleted_at IS NULL
	  AND gam.is_active = TRUE
	UNION ALL
	SELECT gom.group_id, 'organization'::text AS membership_source, NULL::text AS membership_role, 2 AS priority
	FROM iam.group_organization_members AS gom
	JOIN iam.memberships AS m ON m.organization_id = gom.organization_id
	WHERE m.account_id = @account_id
	  AND m.deleted_at IS NULL
	  AND m.is_active = TRUE
	  AND gom.deleted_at IS NULL
	  AND gom.is_active = TRUE
)
SELECT DISTINCT ON (g.id)
	g.id,
	g.name,
	g.slug,
	g.description,
	g.is_active,
	g.created_at,
	ag.membership_source,
	ag.membership_role
FROM accessible_groups AS ag
JOIN iam.groups AS g ON g.id = ag.group_id
WHERE g.is_active = TRUE
ORDER BY g.id, ag.priority ASC, g.created_at DESC
`

	rows := make([]row, 0)
	if err := r.dbFrom(ctx).WithContext(ctx).Raw(query, map[string]any{"account_id": accountID.UUID()}).Scan(&rows).Error; err != nil {
		return nil, err
	}

	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].CreatedAt.After(rows[j].CreatedAt)
	})

	out := make([]ports.GroupMembershipView, 0, len(rows))
	for _, item := range rows {
		out = append(out, ports.GroupMembershipView{
			ID:               item.ID,
			Name:             item.Name,
			Slug:             item.Slug,
			Description:      item.Description,
			IsActive:         item.IsActive,
			CreatedAt:        item.CreatedAt,
			MembershipSource: item.MembershipSource,
			MembershipRole:   item.MembershipRole,
		})
	}
	return out, nil
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
