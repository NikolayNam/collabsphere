package postgres

import (
	"context"
	"fmt"

	"github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres/dbmodel"

	apperrors "github.com/NikolayNam/collabsphere/internal/memberships/application/errors"
	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
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
		OrganizationID: organizationID.UUID(),
		AccountID:      m.AccountID().UUID(),
		Kind:           string(m.Kind()),
		Status:         string(m.Status()),
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
