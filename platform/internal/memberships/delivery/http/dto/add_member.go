package dto

import (
	"time"

	"github.com/google/uuid"
)

type AddMemberInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
	Body           struct {
		AccountID string `json:"accountId" required:"true" format:"uuid"`
		Role      string `json:"role,omitempty" enum:"owner,admin,manager,member,viewer"`
	}
}

type UpdateMemberInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
	MembershipID   string `path:"membership_id" required:"true" format:"uuid"`
	Body           struct {
		Role     *string `json:"role,omitempty" enum:"owner,admin,manager,member,viewer"`
		IsActive *bool   `json:"isActive,omitempty"`
	}
}

type RemoveMemberInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
	MembershipID   string `path:"membership_id" required:"true" format:"uuid"`
}

type ListMembersInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
}

type MemberPayload struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organizationId"`
	AccountID      uuid.UUID  `json:"accountId"`
	Role           string     `json:"role"`
	IsActive       bool       `json:"isActive"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
	DeletedAt      *time.Time `json:"deletedAt,omitempty"`
}

type MemberResponse struct {
	Status int
	Body   MemberPayload
}

type MembersListResponse struct {
	Status int
	Body   struct {
		Members []MemberPayload `json:"members"`
	}
}

type EmptyResponse struct {
	Status int
}
