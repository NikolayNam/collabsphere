package dto

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationBody struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	LogoObjectID *uuid.UUID `json:"logoObjectId,omitempty"`
	Description  *string    `json:"description,omitempty"`
	Website      *string    `json:"website,omitempty"`
	PrimaryEmail *string    `json:"primaryEmail,omitempty"`
	Phone        *string    `json:"phone,omitempty"`
	Address      *string    `json:"address,omitempty"`
	Industry     *string    `json:"industry,omitempty"`
	IsActive     bool       `json:"isActive"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
}

type OrganizationResponse struct {
	Status int              `json:"-"`
	Body   OrganizationBody `json:"body"`
}

type MemberBody struct {
	ID        uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"accountId"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"created_at"`
}
type MembersResponse struct {
	OrganizationID uuid.UUID  `json:"organizationId"`
	Status         int        `json:"-"`
	Body           MemberBody `json:"body"`
}

type MembersListBody struct {
	Data []MemberBody `json:"data"`
}

type MembersListResponse struct {
	Status int             `json:"-"`
	Body   MembersListBody `json:"-"`
}
