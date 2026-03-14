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

type CreateInvitationInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
	Body           struct {
		Email string `json:"email" required:"true" format:"email"`
		Role  string `json:"role,omitempty" enum:"owner,admin,manager,member,viewer"`
	}
}

type ListInvitationsInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
}

type AcceptInvitationInput struct {
	Token string `path:"token" required:"true"`
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

type InvitationPayload struct {
	ID                  uuid.UUID  `json:"id"`
	OrganizationID      uuid.UUID  `json:"organizationId"`
	Email               string     `json:"email"`
	Role                string     `json:"role"`
	Status              string     `json:"status"`
	Token               *string    `json:"token,omitempty"`
	InviterAccountID    uuid.UUID  `json:"inviterAccountId"`
	AcceptedByAccountID *uuid.UUID `json:"acceptedByAccountId,omitempty"`
	AcceptedAt          *time.Time `json:"acceptedAt,omitempty"`
	ExpiresAt           time.Time  `json:"expiresAt"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           *time.Time `json:"updatedAt,omitempty"`
}

type InvitationResponse struct {
	Status int
	Body   InvitationPayload
}

type InvitationsListResponse struct {
	Status int
	Body   struct {
		Invitations []InvitationPayload `json:"invitations"`
	}
}

type AcceptInvitationResponse struct {
	Status int
	Body   struct {
		Invitation InvitationPayload `json:"invitation"`
		Member     MemberPayload     `json:"member"`
	}
}

type EmptyResponse struct {
	Status int
}
