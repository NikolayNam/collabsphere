package postgres

import (
    "context"
    "fmt"

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

    updatedAt := m.CreatedAt()
    if m.UpdatedAt() != nil {
        updatedAt = *m.UpdatedAt()
    }

    mm := &dbmodel.Membership{
        ID:             m.ID(),
        OrganizationID: organizationID.UUID(),
        AccountID:      m.AccountID().UUID(),
        Role:           string(m.Role()),
        IsActive:       m.IsActive(),
        CreatedAt:      m.CreatedAt(),
        UpdatedAt:      updatedAt,
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
