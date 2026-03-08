package postgres

import (
	"context"
	"errors"
	"time"

	accDomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	"github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres/dbmodel"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *MembershipRepo) GetMemberByAccount(ctx context.Context, orgID orgDomain.OrganizationID, accountID accDomain.AccountID) (*memberDomain.Membership, error) {
	var model dbmodel.Membership
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.memberships").
		Where("organization_id = ? AND account_id = ?", orgID.UUID(), accountID.UUID()).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapMembershipModel(model)
}

func (r *MembershipRepo) GetMemberByID(ctx context.Context, orgID orgDomain.OrganizationID, membershipID uuid.UUID) (*memberDomain.Membership, error) {
	var model dbmodel.Membership
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.memberships").
		Where("organization_id = ? AND id = ?", orgID.UUID(), membershipID).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapMembershipModel(model)
}

func (r *MembershipRepo) CountActiveMembersByRole(ctx context.Context, orgID orgDomain.OrganizationID, role memberDomain.MembershipRole) (int64, error) {
	var count int64
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.memberships").
		Where("organization_id = ? AND role = ? AND is_active = ? AND deleted_at IS NULL", orgID.UUID(), string(role), true).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *MembershipRepo) ListMembers(ctx context.Context, orgID orgDomain.OrganizationID) ([]memberDomain.MemberView, error) {
	type row struct {
		ID             uuid.UUID
		OrganizationID uuid.UUID
		AccountID      uuid.UUID
		Role           string
		IsActive       bool
		CreatedAt      time.Time
		UpdatedAt      *time.Time
		DeletedAt      *time.Time
	}

	var rows []row
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.memberships").
		Select("id, organization_id, account_id, role, is_active, created_at, updated_at, deleted_at").
		Where("organization_id = ?", orgID.UUID()).
		Order("created_at asc").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]memberDomain.MemberView, 0, len(rows))
	for _, row := range rows {
		out = append(out, memberDomain.MemberView{
			MembershipID:   row.ID,
			OrganizationID: row.OrganizationID,
			AccountID:      row.AccountID,
			Role:           row.Role,
			IsActive:       row.IsActive,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      cloneTimePtr(row.UpdatedAt),
			DeletedAt:      cloneTimePtr(row.DeletedAt),
		})
	}
	return out, nil
}

func mapMembershipModel(model dbmodel.Membership) (*memberDomain.Membership, error) {
	orgID, err := orgDomain.OrganizationIDFromUUID(model.OrganizationID)
	if err != nil {
		return nil, err
	}
	accountID, err := accDomain.AccountIDFromUUID(model.AccountID)
	if err != nil {
		return nil, err
	}
	return memberDomain.RehydrateMembership(memberDomain.RehydrateMembershipParams{
		ID:             model.ID,
		OrganizationID: orgID,
		AccountID:      accountID,
		Role:           memberDomain.ParseMembershipRole(model.Role),
		IsActive:       model.IsActive,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      cloneTimePtr(&model.UpdatedAt),
		DeletedAt:      cloneTimePtr(model.DeletedAt),
	})
}

func cloneTimePtr(in *time.Time) *time.Time {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}
