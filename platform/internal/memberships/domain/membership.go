package domain

import (
	"strings"
	"time"

	acc "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	org "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type MembershipRole string

const (
	MembershipRoleOwner   MembershipRole = "owner"
	MembershipRoleAdmin   MembershipRole = "admin"
	MembershipRoleManager MembershipRole = "manager"
	MembershipRoleMember  MembershipRole = "member"
	MembershipRoleViewer  MembershipRole = "viewer"
)

func ParseMembershipRole(raw string) MembershipRole {
	return MembershipRole(strings.ToLower(strings.TrimSpace(raw)))
}

func (r MembershipRole) IsValid() bool {
	switch r {
	case MembershipRoleOwner, MembershipRoleAdmin, MembershipRoleManager, MembershipRoleMember, MembershipRoleViewer:
		return true
	default:
		return false
	}
}

func (r MembershipRole) CanManageMembers() bool {
	return r == MembershipRoleOwner || r == MembershipRoleAdmin
}

func (r MembershipRole) CanManageOrganizationProfile() bool {
	return r == MembershipRoleOwner || r == MembershipRoleAdmin
}

func (r MembershipRole) CanManageCatalog() bool {
	return r == MembershipRoleOwner || r == MembershipRoleAdmin || r == MembershipRoleManager
}

func (r MembershipRole) CanAssign(target MembershipRole) bool {
	switch r {
	case MembershipRoleOwner:
		return target.IsValid()
	case MembershipRoleAdmin:
		return target == MembershipRoleManager || target == MembershipRoleMember || target == MembershipRoleViewer
	default:
		return false
	}
}

func (r MembershipRole) CanManageTarget(target MembershipRole) bool {
	switch r {
	case MembershipRoleOwner:
		return target.IsValid()
	case MembershipRoleAdmin:
		return target == MembershipRoleManager || target == MembershipRoleMember || target == MembershipRoleViewer
	default:
		return false
	}
}

type Membership struct {
	id             uuid.UUID
	organizationID org.OrganizationID
	accountID      acc.AccountID
	role           MembershipRole
	isActive       bool
	createdAt      time.Time
	updatedAt      *time.Time
	deletedAt      *time.Time
}

type NewMembershipParams struct {
	OrganizationID org.OrganizationID
	AccountID      acc.AccountID
	Role           MembershipRole
	Now            time.Time
}

type RehydrateMembershipParams struct {
	ID             uuid.UUID
	OrganizationID org.OrganizationID
	AccountID      acc.AccountID
	Role           MembershipRole
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
}

func NewMembership(p NewMembershipParams) (*Membership, error) {
	if p.OrganizationID.IsZero() || p.AccountID.IsZero() {
		return nil, ErrMembershipInvalid
	}
	if strings.TrimSpace(string(p.Role)) == "" {
		return nil, ErrMembershipInvalid
	}
	if p.Now.IsZero() {
		return nil, ErrNowRequired
	}

	updatedAt := p.Now
	return &Membership{
		id:             uuid.New(),
		organizationID: p.OrganizationID,
		accountID:      p.AccountID,
		role:           p.Role,
		isActive:       true,
		createdAt:      p.Now,
		updatedAt:      &updatedAt,
	}, nil
}

func RehydrateMembership(p RehydrateMembershipParams) (*Membership, error) {
	if p.ID == uuid.Nil || p.OrganizationID.IsZero() || p.AccountID.IsZero() {
		return nil, ErrMembershipInvalid
	}
	if !p.Role.IsValid() {
		return nil, ErrMembershipInvalid
	}
	if p.CreatedAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	if p.UpdatedAt != nil && p.UpdatedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}
	if p.DeletedAt != nil && p.DeletedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}
	return &Membership{
		id:             p.ID,
		organizationID: p.OrganizationID,
		accountID:      p.AccountID,
		role:           p.Role,
		isActive:       p.IsActive,
		createdAt:      p.CreatedAt,
		updatedAt:      cloneTimePtr(p.UpdatedAt),
		deletedAt:      cloneTimePtr(p.DeletedAt),
	}, nil
}

func (m *Membership) ID() uuid.UUID                      { return m.id }
func (m *Membership) OrganizationID() org.OrganizationID { return m.organizationID }
func (m *Membership) AccountID() acc.AccountID           { return m.accountID }
func (m *Membership) Role() MembershipRole               { return m.role }
func (m *Membership) IsActive() bool                     { return m.isActive }
func (m *Membership) IsRemoved() bool                    { return m.deletedAt != nil }
func (m *Membership) CreatedAt() time.Time               { return m.createdAt }
func (m *Membership) UpdatedAt() *time.Time              { return cloneTimePtr(m.updatedAt) }
func (m *Membership) DeletedAt() *time.Time              { return cloneTimePtr(m.deletedAt) }

func (m *Membership) ChangeRole(role MembershipRole, now time.Time) error {
	if strings.TrimSpace(string(role)) == "" || now.IsZero() {
		return ErrMembershipInvalid
	}
	m.role = role
	m.updatedAt = &now
	return nil
}

func (m *Membership) Activate(now time.Time) error {
	if now.IsZero() {
		return ErrNowRequired
	}
	m.isActive = true
	m.deletedAt = nil
	m.updatedAt = &now
	return nil
}

func (m *Membership) Suspend(now time.Time) error {
	if now.IsZero() {
		return ErrNowRequired
	}
	m.isActive = false
	m.updatedAt = &now
	return nil
}

func (m *Membership) Remove(now time.Time) error {
	if now.IsZero() {
		return ErrNowRequired
	}
	m.isActive = false
	m.deletedAt = &now
	m.updatedAt = &now
	return nil
}

func cloneTimePtr(in *time.Time) *time.Time {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}
