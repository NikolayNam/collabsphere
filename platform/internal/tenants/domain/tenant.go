package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type TenantRole string

const (
	TenantRoleOwner  TenantRole = "owner"
	TenantRoleAdmin  TenantRole = "admin"
	TenantRoleMember TenantRole = "member"
)

type Tenant struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

type TenantMember struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	AccountID uuid.UUID
	Role      TenantRole
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

type TenantOrganization struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	OrganizationID uuid.UUID
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
}

type TenantMembershipView struct {
	ID             uuid.UUID
	Name           string
	Slug           string
	Description    *string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	MembershipRole TenantRole
}

func NormalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	v := trimmed
	return &v
}

func NormalizeSlug(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func (r TenantRole) IsValid() bool {
	return r == TenantRoleOwner || r == TenantRoleAdmin || r == TenantRoleMember
}

func (r TenantRole) CanManage() bool {
	return r == TenantRoleOwner || r == TenantRoleAdmin
}
