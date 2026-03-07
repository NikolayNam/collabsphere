package domain

import (
    "time"

    acc "github.com/NikolayNam/collabsphere/internal/accounts/domain"
    org "github.com/NikolayNam/collabsphere/internal/organizations/domain"
    "github.com/google/uuid"
)

type MembershipRole string

const (
    MembershipRoleOwner  MembershipRole = "owner"
    MembershipRoleMember MembershipRole = "member"
)

func (r MembershipRole) IsValid() bool {
    return r == MembershipRoleOwner || r == MembershipRoleMember
}

type Membership struct {
    id             uuid.UUID
    organizationID org.OrganizationID
    accountID      acc.AccountID
    role           MembershipRole
    isActive       bool
    createdAt      time.Time
    updatedAt      *time.Time
}

type NewMembershipParams struct {
    OrganizationID org.OrganizationID
    AccountID      acc.AccountID
    Role           MembershipRole
    Now            time.Time
}

func NewMembership(p NewMembershipParams) (*Membership, error) {
    if p.OrganizationID.IsZero() || p.AccountID.IsZero() {
        return nil, ErrMembershipInvalid
    }
    if !p.Role.IsValid() {
        return nil, ErrMembershipInvalid
    }
    if p.Now.IsZero() {
        return nil, ErrNowRequired
    }

    id := uuid.New()
    updatedAt := p.Now

    return &Membership{
        id:             id,
        organizationID: p.OrganizationID,
        accountID:      p.AccountID,
        role:           p.Role,
        isActive:       true,
        createdAt:      p.Now,
        updatedAt:      &updatedAt,
    }, nil
}

func (m *Membership) ID() uuid.UUID {
    return m.id
}

func (m *Membership) OrganizationID() org.OrganizationID {
    return m.organizationID
}

func (m *Membership) AccountID() acc.AccountID {
    return m.accountID
}

func (m *Membership) Role() MembershipRole {
    return m.role
}

func (m *Membership) IsActive() bool {
    return m.isActive
}

func (m *Membership) CreatedAt() time.Time {
    return m.createdAt
}

func (m *Membership) UpdatedAt() *time.Time {
    if m.updatedAt == nil {
        return nil
    }
    v := *m.updatedAt
    return &v
}
