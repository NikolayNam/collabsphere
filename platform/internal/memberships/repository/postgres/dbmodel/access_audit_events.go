package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationAccessAuditEvent struct {
	ID               uuid.UUID  `gorm:"column:id"`
	OrganizationID   uuid.UUID  `gorm:"column:organization_id"`
	ActorSubjectType string     `gorm:"column:actor_subject_type"`
	ActorSubjectID   *uuid.UUID `gorm:"column:actor_subject_id"`
	ActorAccountID   *uuid.UUID `gorm:"column:actor_account_id"`
	Action           string     `gorm:"column:action"`
	TargetType       string     `gorm:"column:target_type"`
	TargetID         *uuid.UUID `gorm:"column:target_id"`
	RequestID        *string    `gorm:"column:request_id"`
	PreviousState    []byte     `gorm:"column:previous_state_json;type:jsonb"`
	NextState        []byte     `gorm:"column:next_state_json;type:jsonb"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
}

func (OrganizationAccessAuditEvent) TableName() string {
	return "iam.organization_access_audit_events"
}
