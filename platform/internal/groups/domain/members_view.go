package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccountMemberView struct {
	MembershipID uuid.UUID
	AccountID    uuid.UUID
	Email        string
	DisplayName  *string
	Role         string
	IsActive     bool
	CreatedAt    time.Time
}

type OrganizationMemberView struct {
	MembershipID   uuid.UUID
	OrganizationID uuid.UUID
	Name           string
	Slug           string
	IsActive       bool
	CreatedAt      time.Time
}

type MembersView struct {
	Accounts      []AccountMemberView
	Organizations []OrganizationMemberView
}
