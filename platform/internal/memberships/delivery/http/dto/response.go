package dto

import (
	"time"

	"github.com/google/uuid"
)

type MemberBody struct {
	ID        uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"accountId"`
	Kind      string    `json:"kind"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
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

//accID: 08c7b644-57b7-4971-bf24-6714cf9ccd17

// org93c0a705-e922-47f6-8ecb-d3d15547bd6c
// 0eed5c40-5df9-42f0-8cd7-b1e155364524
