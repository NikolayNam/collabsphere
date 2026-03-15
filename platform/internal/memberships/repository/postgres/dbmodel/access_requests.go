package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationAccessRequest struct {
	ID                 uuid.UUID  `gorm:"column:id"`
	OrganizationID     uuid.UUID  `gorm:"column:organization_id"`
	RequesterAccountID uuid.UUID  `gorm:"column:requester_account_id"`
	RequestedRole      string     `gorm:"column:requested_role"`
	Message            *string    `gorm:"column:message"`
	Status             string     `gorm:"column:status"`
	ReviewerAccountID  *uuid.UUID `gorm:"column:reviewer_account_id"`
	ReviewNote         *string    `gorm:"column:review_note"`
	ReviewedAt         *time.Time `gorm:"column:reviewed_at"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
}

func (OrganizationAccessRequest) TableName() string {
	return "iam.organization_access_requests"
}
