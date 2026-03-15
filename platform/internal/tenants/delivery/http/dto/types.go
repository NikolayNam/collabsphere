package dto

import (
	"time"

	"github.com/google/uuid"
)

type TenantPayload struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description,omitempty"`
	IsActive    bool       `json:"isActive"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

type TenantResponse struct {
	Status int           `json:"-"`
	Body   TenantPayload `json:"body"`
}

type TenantMemberPayload struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenantId"`
	AccountID uuid.UUID  `json:"accountId"`
	Role      string     `json:"role"`
	IsActive  bool       `json:"isActive"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

type TenantOrganizationPayload struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenantId"`
	OrganizationID uuid.UUID  `json:"organizationId"`
	IsActive       bool       `json:"isActive"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
	DeletedAt      *time.Time `json:"deletedAt,omitempty"`
}

type TenantMembersListResponse struct {
	Status int `json:"-"`
	Body   struct {
		Members []TenantMemberPayload `json:"members"`
	} `json:"body"`
}

type TenantOrganizationsListResponse struct {
	Status int `json:"-"`
	Body   struct {
		Organizations []TenantOrganizationPayload `json:"organizations"`
	} `json:"body"`
}

type MyTenantsResponse struct {
	Status int `json:"-"`
	Body   struct {
		Data []MyTenantPayload `json:"data"`
	} `json:"body"`
}

type MyTenantPayload struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Slug           string     `json:"slug"`
	Description    *string    `json:"description,omitempty"`
	IsActive       bool       `json:"isActive"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
	MembershipRole string     `json:"membershipRole"`
}

type EmptyResponse struct {
	Status int `json:"-"`
}
