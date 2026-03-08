package postgres

import (
	"context"
	"fmt"
	"time"

	apperrors "github.com/NikolayNam/collabsphere/internal/memberships/application/errors"
	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	"github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres/dbmodel"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

func (r *MembershipRepo) AddMember(ctx context.Context, organizationID orgDomain.OrganizationID, m *memberDomain.Membership) error {
	if organizationID.IsZero() {
		return apperrors.InvalidInput("Invalid organization_id")
	}
	if m == nil {
		return apperrors.InvalidInput("Membership is required")
	}

	mm := &dbmodel.Membership{
		ID:             m.ID(),
		OrganizationID: organizationID.UUID(),
		AccountID:      m.AccountID().UUID(),
		Role:           string(m.Role()),
		IsActive:       m.IsActive(),
		CreatedAt:      m.CreatedAt(),
		UpdatedAt:      derefTime(m.UpdatedAt(), m.CreatedAt()),
		DeletedAt:      m.DeletedAt(),
	}

	err := r.dbFrom(ctx).WithContext(ctx).Create(mm).Error
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.MemberAlreadyExists()
		}
		if isForeignKeyViolation(err) {
			return apperrors.InvalidInput("Organization or account not found")
		}
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (r *MembershipRepo) SaveMember(ctx context.Context, organizationID orgDomain.OrganizationID, m *memberDomain.Membership) error {
	if organizationID.IsZero() {
		return apperrors.InvalidInput("Invalid organization_id")
	}
	if m == nil {
		return apperrors.InvalidInput("Membership is required")
	}

	updates := map[string]any{
		"role":       string(m.Role()),
		"is_active":  m.IsActive(),
		"updated_at": derefTime(m.UpdatedAt(), m.CreatedAt()),
		"deleted_at": m.DeletedAt(),
	}

	err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.memberships").
		Where("id = ? AND organization_id = ?", m.ID(), organizationID.UUID()).
		Updates(updates).Error
	if err != nil {
		return fmt.Errorf("save member: %w", err)
	}
	return nil
}

func derefTime(value *time.Time, fallback time.Time) time.Time {
	if value == nil {
		return fallback
	}
	return *value
}
