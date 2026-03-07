package dto

import (
	"time"

	"github.com/google/uuid"
)

type GroupBody struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
}

type GroupResponse struct {
	Status int       `json:"-"`
	Body   GroupBody `json:"body"`
}

type GroupAccountMemberBody struct {
	ID          uuid.UUID `json:"id"`
	AccountID   uuid.UUID `json:"accountId"`
	Email       string    `json:"email"`
	DisplayName *string   `json:"displayName,omitempty"`
	Role        string    `json:"role"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
}

type GroupOrganizationMemberBody struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organizationId"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	IsActive       bool      `json:"isActive"`
	CreatedAt      time.Time `json:"createdAt"`
}

type GroupAccountMemberResponse struct {
	Status int                    `json:"-"`
	Body   GroupAccountMemberBody `json:"body"`
}

type GroupOrganizationMemberResponse struct {
	Status int                         `json:"-"`
	Body   GroupOrganizationMemberBody `json:"body"`
}

type GroupMembersBody struct {
	Accounts      []GroupAccountMemberBody      `json:"accounts"`
	Organizations []GroupOrganizationMemberBody `json:"organizations"`
}

type GroupMembersResponse struct {
	Status int              `json:"-"`
	Body   GroupMembersBody `json:"body"`
}
