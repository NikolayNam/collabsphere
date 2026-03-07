package domain

import (
	"time"

	acc "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	org "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type MembershipKind string

const (
	MembershipKindOwner  MembershipKind = "owner"
	MembershipKindMember MembershipKind = "member"
)

func (k MembershipKind) IsValid() bool {
	return k == MembershipKindOwner || k == MembershipKindMember
}

type MembershipStatus string

const (
	MembershipStatusActive  MembershipStatus = "active"
	MembershipStatusInvited MembershipStatus = "invited"
	MembershipStatusRemoved MembershipStatus = "removed"
)

func (s MembershipStatus) IsValid() bool {
	switch s {
	case MembershipStatusActive, MembershipStatusInvited, MembershipStatusRemoved:
		return true
	default:
		return false
	}
}

type Membership struct {
	id             MembershipID
	organizationID org.OrganizationID
	accountID      acc.AccountID

	kind   MembershipKind
	status MembershipStatus

	createdAt time.Time
	updatedAt *time.Time
}

type MembershipID [16]byte

type NewMembershipParams struct {
	OrganizationID org.OrganizationID
	AccountID      acc.AccountID
	Kind           MembershipKind
	Status         MembershipStatus
	Now            time.Time
}

func NewMembership(p NewMembershipParams) (*Membership, error) {
	if p.OrganizationID.IsZero() || p.AccountID.IsZero() {
		return nil, ErrMembershipInvalid
	}
	if !p.Kind.IsValid() || !p.Status.IsValid() {
		return nil, ErrMembershipInvalid
	}
	if p.Now.IsZero() {
		return nil, ErrNowRequired
	}

	id := uuid.New()
	var mid MembershipID
	copy(mid[:], id[:])

	return &Membership{
		id:             mid,
		organizationID: p.OrganizationID,
		accountID:      p.AccountID,
		kind:           p.Kind,
		status:         p.Status,
		createdAt:      p.Now,
		updatedAt:      nil,
	}, nil
}

func (m *Membership) ID() uuid.UUID {
	return uuid.UUID(m.id)
}

func (m *Membership) OrganizationID() org.OrganizationID {
	return m.organizationID
}

func (m *Membership) AccountID() acc.AccountID {
	return m.accountID
}

func (m *Membership) Kind() MembershipKind {
	return m.kind
}

func (m *Membership) Status() MembershipStatus {
	return m.status
}

func (m *Membership) CreatedAt() time.Time {
	return m.createdAt
}

func (m *Membership) UpdatedAt() *time.Time {
	if m.updatedAt == nil {
		return nil
	}
	return new(*m.updatedAt)
}
