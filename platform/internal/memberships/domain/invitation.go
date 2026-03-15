package domain

import (
	"strings"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusExpired  InvitationStatus = "expired"
	InvitationStatusRevoked  InvitationStatus = "revoked"
)

type OrganizationInvitation struct {
	id                  uuid.UUID
	organizationID      orgdomain.OrganizationID
	email               accdomain.Email
	role                MembershipRole
	tokenHash           string
	inviterAccountID    uuid.UUID
	acceptedByAccountID *uuid.UUID
	acceptedAt          *time.Time
	revokedByAccountID  *uuid.UUID
	revokedAt           *time.Time
	expiresAt           time.Time
	createdAt           time.Time
	updatedAt           *time.Time
}

type NewOrganizationInvitationParams struct {
	OrganizationID   orgdomain.OrganizationID
	Email            accdomain.Email
	Role             MembershipRole
	TokenHash        string
	InviterAccountID uuid.UUID
	ExpiresAt        time.Time
	Now              time.Time
}

type RehydrateOrganizationInvitationParams struct {
	ID                  uuid.UUID
	OrganizationID      orgdomain.OrganizationID
	Email               accdomain.Email
	Role                MembershipRole
	TokenHash           string
	InviterAccountID    uuid.UUID
	AcceptedByAccountID *uuid.UUID
	AcceptedAt          *time.Time
	RevokedByAccountID  *uuid.UUID
	RevokedAt           *time.Time
	ExpiresAt           time.Time
	CreatedAt           time.Time
	UpdatedAt           *time.Time
}

type InvitationView struct {
	ID                  uuid.UUID
	OrganizationID      uuid.UUID
	Email               string
	Role                string
	Status              string
	InviterAccountID    uuid.UUID
	AcceptedByAccountID *uuid.UUID
	AcceptedAt          *time.Time
	ExpiresAt           time.Time
	CreatedAt           time.Time
	UpdatedAt           *time.Time
}

func NewOrganizationInvitation(p NewOrganizationInvitationParams) (*OrganizationInvitation, error) {
	if p.OrganizationID.IsZero() || p.Email.IsZero() || strings.TrimSpace(string(p.Role)) == "" || p.InviterAccountID == uuid.Nil || p.TokenHash == "" {
		return nil, ErrMembershipInvalid
	}
	if p.Now.IsZero() || p.ExpiresAt.IsZero() || !p.ExpiresAt.After(p.Now) {
		return nil, ErrTimestampsInvalid
	}
	updatedAt := p.Now
	return &OrganizationInvitation{
		id:               uuid.New(),
		organizationID:   p.OrganizationID,
		email:            p.Email,
		role:             p.Role,
		tokenHash:        p.TokenHash,
		inviterAccountID: p.InviterAccountID,
		expiresAt:        p.ExpiresAt,
		createdAt:        p.Now,
		updatedAt:        &updatedAt,
	}, nil
}

func RehydrateOrganizationInvitation(p RehydrateOrganizationInvitationParams) (*OrganizationInvitation, error) {
	if p.ID == uuid.Nil || p.OrganizationID.IsZero() || p.Email.IsZero() || strings.TrimSpace(string(p.Role)) == "" || p.InviterAccountID == uuid.Nil || p.TokenHash == "" {
		return nil, ErrMembershipInvalid
	}
	if p.CreatedAt.IsZero() || p.ExpiresAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	return &OrganizationInvitation{
		id:                  p.ID,
		organizationID:      p.OrganizationID,
		email:               p.Email,
		role:                p.Role,
		tokenHash:           p.TokenHash,
		inviterAccountID:    p.InviterAccountID,
		acceptedByAccountID: cloneUUIDPtr(p.AcceptedByAccountID),
		acceptedAt:          cloneTimePtr(p.AcceptedAt),
		revokedByAccountID:  cloneUUIDPtr(p.RevokedByAccountID),
		revokedAt:           cloneTimePtr(p.RevokedAt),
		expiresAt:           p.ExpiresAt,
		createdAt:           p.CreatedAt,
		updatedAt:           cloneTimePtr(p.UpdatedAt),
	}, nil
}

func (i *OrganizationInvitation) ID() uuid.UUID                            { return i.id }
func (i *OrganizationInvitation) OrganizationID() orgdomain.OrganizationID { return i.organizationID }
func (i *OrganizationInvitation) Email() accdomain.Email                   { return i.email }
func (i *OrganizationInvitation) Role() MembershipRole                     { return i.role }
func (i *OrganizationInvitation) TokenHash() string                        { return i.tokenHash }
func (i *OrganizationInvitation) InviterAccountID() uuid.UUID              { return i.inviterAccountID }
func (i *OrganizationInvitation) AcceptedByAccountID() *uuid.UUID {
	return cloneUUIDPtr(i.acceptedByAccountID)
}
func (i *OrganizationInvitation) AcceptedAt() *time.Time { return cloneTimePtr(i.acceptedAt) }
func (i *OrganizationInvitation) RevokedByAccountID() *uuid.UUID {
	return cloneUUIDPtr(i.revokedByAccountID)
}
func (i *OrganizationInvitation) RevokedAt() *time.Time { return cloneTimePtr(i.revokedAt) }
func (i *OrganizationInvitation) ExpiresAt() time.Time  { return i.expiresAt }
func (i *OrganizationInvitation) CreatedAt() time.Time  { return i.createdAt }
func (i *OrganizationInvitation) UpdatedAt() *time.Time { return cloneTimePtr(i.updatedAt) }

func (i *OrganizationInvitation) Status(now time.Time) InvitationStatus {
	switch {
	case i.acceptedAt != nil:
		return InvitationStatusAccepted
	case i.revokedAt != nil:
		return InvitationStatusRevoked
	case !i.expiresAt.After(now):
		return InvitationStatusExpired
	default:
		return InvitationStatusPending
	}
}

func (i *OrganizationInvitation) Accept(accountID uuid.UUID, now time.Time) error {
	if accountID == uuid.Nil || now.IsZero() {
		return ErrMembershipInvalid
	}
	if i.Status(now) != InvitationStatusPending {
		return ErrMembershipInvalid
	}
	i.acceptedByAccountID = &accountID
	i.acceptedAt = &now
	i.updatedAt = &now
	return nil
}

func (i *OrganizationInvitation) Revoke(actorAccountID uuid.UUID, now time.Time) error {
	if actorAccountID == uuid.Nil || now.IsZero() {
		return ErrMembershipInvalid
	}
	if i.acceptedAt != nil || i.revokedAt != nil {
		return ErrMembershipInvalid
	}
	i.revokedByAccountID = &actorAccountID
	i.revokedAt = &now
	i.updatedAt = &now
	return nil
}

func (i *OrganizationInvitation) ToView(now time.Time) InvitationView {
	return InvitationView{
		ID:                  i.id,
		OrganizationID:      i.organizationID.UUID(),
		Email:               i.email.String(),
		Role:                string(i.role),
		Status:              string(i.Status(now)),
		InviterAccountID:    i.inviterAccountID,
		AcceptedByAccountID: cloneUUIDPtr(i.acceptedByAccountID),
		AcceptedAt:          cloneTimePtr(i.acceptedAt),
		ExpiresAt:           i.expiresAt,
		CreatedAt:           i.createdAt,
		UpdatedAt:           cloneTimePtr(i.updatedAt),
	}
}

func cloneUUIDPtr(in *uuid.UUID) *uuid.UUID {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}
