package dto

import (
	"time"

	"github.com/google/uuid"
)

type AddMemberInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
	Body           struct {
		AccountID string `json:"accountId" required:"true" format:"uuid"`
		Role      string `json:"role,omitempty" doc:"System role (owner, admin, manager, member, viewer) or custom role code"`
	}
}

type UpdateMemberInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
	MembershipID   string `path:"membership_id" required:"true" format:"uuid"`
	Body           struct {
		Role     *string `json:"role,omitempty" doc:"System or custom role code"`
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
		Role  string `json:"role,omitempty" doc:"System or custom role code"`
	}
}

type ListInvitationsInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
}

type AcceptInvitationInput struct {
	Token string `path:"token" required:"true"`
}

type CreateAccessRequestInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
	Body           struct {
		Role    string  `json:"role,omitempty" doc:"System or custom role code"`
		Message *string `json:"message,omitempty"`
	}
}

type ListAccessRequestsInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
}

type ReviewAccessRequestInput struct {
	OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
	RequestID      string `path:"request_id" required:"true" format:"uuid"`
	Body           struct {
		ReviewNote *string `json:"reviewNote,omitempty"`
	}
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

type AccessRequestPayload struct {
	ID               uuid.UUID  `json:"id"`
	OrganizationID   uuid.UUID  `json:"organizationId"`
	RequesterAccount uuid.UUID  `json:"requesterAccountId"`
	RequestedRole    string     `json:"requestedRole"`
	Message          *string    `json:"message,omitempty"`
	Status           string     `json:"status"`
	ReviewerAccount  *uuid.UUID `json:"reviewerAccountId,omitempty"`
	ReviewNote       *string    `json:"reviewNote,omitempty"`
	ReviewedAt       *time.Time `json:"reviewedAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        *time.Time `json:"updatedAt,omitempty"`
}

type AccessRequestResponse struct {
	Status int
	Body   AccessRequestPayload
}

type AccessRequestsListResponse struct {
	Status int
	Body   struct {
		Requests []AccessRequestPayload `json:"requests"`
	}
}

type ReviewAccessRequestResponse struct {
	Status int
	Body   struct {
		Request AccessRequestPayload `json:"request"`
		Member  *MemberPayload       `json:"member,omitempty"`
	}
}

type EmptyResponse struct {
	Status int
}
