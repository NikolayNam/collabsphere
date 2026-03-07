package domain

import (
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/google/uuid"
)

type GroupAccountRole string

const (
	GroupAccountRoleOwner  GroupAccountRole = "owner"
	GroupAccountRoleMember GroupAccountRole = "member"
)

func (r GroupAccountRole) IsValid() bool {
	return r == GroupAccountRoleOwner || r == GroupAccountRoleMember
}

type AccountMember struct {
	id        uuid.UUID
	groupID   GroupID
	accountID accdomain.AccountID
	role      GroupAccountRole
	isActive  bool
	createdAt time.Time
	updatedAt *time.Time
}

type NewAccountMemberParams struct {
	GroupID   GroupID
	AccountID accdomain.AccountID
	Role      GroupAccountRole
	Now       time.Time
}

type RehydrateAccountMemberParams struct {
	ID        uuid.UUID
	GroupID   GroupID
	AccountID accdomain.AccountID
	Role      GroupAccountRole
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewAccountMember(p NewAccountMemberParams) (*AccountMember, error) {
	if p.GroupID.IsZero() || p.AccountID.IsZero() || !p.Role.IsValid() {
		return nil, ErrAccountMemberInvalid
	}
	if p.Now.IsZero() {
		return nil, ErrNowRequired
	}

	updatedAt := p.Now
	return &AccountMember{
		id:        uuid.New(),
		groupID:   p.GroupID,
		accountID: p.AccountID,
		role:      p.Role,
		isActive:  true,
		createdAt: p.Now,
		updatedAt: &updatedAt,
	}, nil
}

func RehydrateAccountMember(p RehydrateAccountMemberParams) (*AccountMember, error) {
	if p.ID == uuid.Nil || p.GroupID.IsZero() || p.AccountID.IsZero() || !p.Role.IsValid() {
		return nil, ErrAccountMemberInvalid
	}
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	if p.UpdatedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}

	updatedAt := p.UpdatedAt
	return &AccountMember{
		id:        p.ID,
		groupID:   p.GroupID,
		accountID: p.AccountID,
		role:      p.Role,
		isActive:  p.IsActive,
		createdAt: p.CreatedAt,
		updatedAt: &updatedAt,
	}, nil
}

func (m *AccountMember) ID() uuid.UUID                  { return m.id }
func (m *AccountMember) GroupID() GroupID               { return m.groupID }
func (m *AccountMember) AccountID() accdomain.AccountID { return m.accountID }
func (m *AccountMember) Role() GroupAccountRole         { return m.role }
func (m *AccountMember) IsActive() bool                 { return m.isActive }
func (m *AccountMember) CreatedAt() time.Time           { return m.createdAt }
func (m *AccountMember) UpdatedAt() *time.Time          { return cloneTimePtr(m.updatedAt) }
