package domain

import (
	"strings"
	"time"

	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type AccessRequestStatus string

const (
	AccessRequestStatusPending  AccessRequestStatus = "pending"
	AccessRequestStatusApproved AccessRequestStatus = "approved"
	AccessRequestStatusRejected AccessRequestStatus = "rejected"
)

type OrganizationAccessRequest struct {
	id               uuid.UUID
	organizationID   orgdomain.OrganizationID
	requesterAccount uuid.UUID
	requestedRole    MembershipRole
	message          *string
	status           AccessRequestStatus
	reviewerAccount  *uuid.UUID
	reviewNote       *string
	reviewedAt       *time.Time
	createdAt        time.Time
	updatedAt        *time.Time
}

type NewOrganizationAccessRequestParams struct {
	OrganizationID   orgdomain.OrganizationID
	RequesterAccount uuid.UUID
	RequestedRole    MembershipRole
	Message          *string
	Now              time.Time
}

type RehydrateOrganizationAccessRequestParams struct {
	ID               uuid.UUID
	OrganizationID   orgdomain.OrganizationID
	RequesterAccount uuid.UUID
	RequestedRole    MembershipRole
	Message          *string
	Status           AccessRequestStatus
	ReviewerAccount  *uuid.UUID
	ReviewNote       *string
	ReviewedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}

type AccessRequestView struct {
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

func NewOrganizationAccessRequest(p NewOrganizationAccessRequestParams) (*OrganizationAccessRequest, error) {
	if p.OrganizationID.IsZero() || p.RequesterAccount == uuid.Nil || strings.TrimSpace(string(p.RequestedRole)) == "" || p.Now.IsZero() {
		return nil, ErrMembershipInvalid
	}
	updated := p.Now
	return &OrganizationAccessRequest{
		id:               uuid.New(),
		organizationID:   p.OrganizationID,
		requesterAccount: p.RequesterAccount,
		requestedRole:    p.RequestedRole,
		message:          normalizeOptionalText(p.Message),
		status:           AccessRequestStatusPending,
		createdAt:        p.Now,
		updatedAt:        &updated,
	}, nil
}

func RehydrateOrganizationAccessRequest(p RehydrateOrganizationAccessRequestParams) (*OrganizationAccessRequest, error) {
	if p.ID == uuid.Nil || p.OrganizationID.IsZero() || p.RequesterAccount == uuid.Nil || strings.TrimSpace(string(p.RequestedRole)) == "" || !p.Status.IsValid() || p.CreatedAt.IsZero() {
		return nil, ErrMembershipInvalid
	}
	return &OrganizationAccessRequest{
		id:               p.ID,
		organizationID:   p.OrganizationID,
		requesterAccount: p.RequesterAccount,
		requestedRole:    p.RequestedRole,
		message:          normalizeOptionalText(p.Message),
		status:           p.Status,
		reviewerAccount:  cloneUUIDPtr(p.ReviewerAccount),
		reviewNote:       normalizeOptionalText(p.ReviewNote),
		reviewedAt:       cloneTimePtr(p.ReviewedAt),
		createdAt:        p.CreatedAt,
		updatedAt:        cloneTimePtr(p.UpdatedAt),
	}, nil
}

func (s AccessRequestStatus) IsValid() bool {
	return s == AccessRequestStatusPending || s == AccessRequestStatusApproved || s == AccessRequestStatusRejected
}

func (r *OrganizationAccessRequest) ID() uuid.UUID { return r.id }
func (r *OrganizationAccessRequest) OrganizationID() orgdomain.OrganizationID {
	return r.organizationID
}
func (r *OrganizationAccessRequest) RequesterAccountID() uuid.UUID { return r.requesterAccount }
func (r *OrganizationAccessRequest) RequestedRole() MembershipRole { return r.requestedRole }
func (r *OrganizationAccessRequest) Message() *string              { return normalizeOptionalText(r.message) }
func (r *OrganizationAccessRequest) Status() AccessRequestStatus   { return r.status }
func (r *OrganizationAccessRequest) ReviewerAccountID() *uuid.UUID {
	return cloneUUIDPtr(r.reviewerAccount)
}
func (r *OrganizationAccessRequest) ReviewNote() *string    { return normalizeOptionalText(r.reviewNote) }
func (r *OrganizationAccessRequest) ReviewedAt() *time.Time { return cloneTimePtr(r.reviewedAt) }
func (r *OrganizationAccessRequest) CreatedAt() time.Time   { return r.createdAt }
func (r *OrganizationAccessRequest) UpdatedAt() *time.Time  { return cloneTimePtr(r.updatedAt) }

func (r *OrganizationAccessRequest) Approve(reviewerAccountID uuid.UUID, reviewNote *string, now time.Time) error {
	if reviewerAccountID == uuid.Nil || now.IsZero() || r.status != AccessRequestStatusPending {
		return ErrMembershipInvalid
	}
	r.status = AccessRequestStatusApproved
	r.reviewerAccount = &reviewerAccountID
	r.reviewNote = normalizeOptionalText(reviewNote)
	r.reviewedAt = &now
	r.updatedAt = &now
	return nil
}

func (r *OrganizationAccessRequest) Reject(reviewerAccountID uuid.UUID, reviewNote *string, now time.Time) error {
	if reviewerAccountID == uuid.Nil || now.IsZero() || r.status != AccessRequestStatusPending {
		return ErrMembershipInvalid
	}
	r.status = AccessRequestStatusRejected
	r.reviewerAccount = &reviewerAccountID
	r.reviewNote = normalizeOptionalText(reviewNote)
	r.reviewedAt = &now
	r.updatedAt = &now
	return nil
}

func (r *OrganizationAccessRequest) ToView() AccessRequestView {
	return AccessRequestView{
		ID:               r.id,
		OrganizationID:   r.organizationID.UUID(),
		RequesterAccount: r.requesterAccount,
		RequestedRole:    string(r.requestedRole),
		Message:          normalizeOptionalText(r.message),
		Status:           string(r.status),
		ReviewerAccount:  cloneUUIDPtr(r.reviewerAccount),
		ReviewNote:       normalizeOptionalText(r.reviewNote),
		ReviewedAt:       cloneTimePtr(r.reviewedAt),
		CreatedAt:        r.createdAt,
		UpdatedAt:        cloneTimePtr(r.updatedAt),
	}
}

func normalizeOptionalText(value *string) *string {
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
