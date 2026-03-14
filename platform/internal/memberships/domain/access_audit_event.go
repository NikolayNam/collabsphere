package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccessAuditEvent struct {
	ID               uuid.UUID
	OrganizationID   uuid.UUID
	ActorSubjectType string
	ActorSubjectID   *uuid.UUID
	ActorAccountID   *uuid.UUID
	Action           string
	TargetType       string
	TargetID         *uuid.UUID
	RequestID        *string
	PreviousState    map[string]any
	NextState        map[string]any
	CreatedAt        time.Time
}
