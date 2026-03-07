package dto

import (
    "time"

    "github.com/google/uuid"
)

type MemberBody struct {
    ID        uuid.UUID `json:"id"`
    AccountID uuid.UUID `json:"accountId"`
    Role      string    `json:"role"`
    IsActive  bool      `json:"isActive"`
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
