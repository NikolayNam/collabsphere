package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationInvitation struct {
	ID                  uuid.UUID  `gorm:"column:id"`
	OrganizationID      uuid.UUID  `gorm:"column:organization_id"`
	Email               string     `gorm:"column:email"`
	Role                string     `gorm:"column:role"`
	TokenHash           string     `gorm:"column:token_hash"`
	InviterAccountID    uuid.UUID  `gorm:"column:inviter_account_id"`
	AcceptedByAccountID *uuid.UUID `gorm:"column:accepted_by_account_id"`
	AcceptedAt          *time.Time `gorm:"column:accepted_at"`
	RevokedByAccountID  *uuid.UUID `gorm:"column:revoked_by_account_id"`
	RevokedAt           *time.Time `gorm:"column:revoked_at"`
	ExpiresAt           time.Time  `gorm:"column:expires_at"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at"`
}

func (OrganizationInvitation) TableName() string {
	return "iam.organization_invitations"
}
