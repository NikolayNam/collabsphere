package postgres

import (
    "context"
    "time"

    "github.com/google/uuid"

    memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
    orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

func (r *MembershipRepo) ListMembers(ctx context.Context, orgID orgDomain.OrganizationID) ([]memberDomain.MemberView, error) {
    type row struct {
        ID             uuid.UUID
        OrganizationID uuid.UUID
        AccountID      uuid.UUID
        Role           string
        IsActive       bool
        CreatedAt      time.Time
    }

    var rows []row
    err := r.dbFrom(ctx).WithContext(ctx).
        Table("iam.memberships").
        Select("id, organization_id, account_id, role, is_active, created_at").
        Where("organization_id = ?", orgID.UUID()).
        Order("created_at asc").
        Scan(&rows).Error
    if err != nil {
        return nil, err
    }

    out := make([]memberDomain.MemberView, 0, len(rows))
    for _, r := range rows {
        out = append(out, memberDomain.MemberView{
            MembershipID:   r.ID,
            OrganizationID: r.OrganizationID,
            AccountID:      r.AccountID,
            Role:           r.Role,
            IsActive:       r.IsActive,
            CreatedAt:      r.CreatedAt,
        })
    }
    return out, nil
}
