package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditStatus string

const (
	AuditStatusSuccess AuditStatus = "success"
	AuditStatusDenied  AuditStatus = "denied"
	AuditStatusFailed  AuditStatus = "failed"
)

type AuditEvent struct {
	ActorAccountID *uuid.UUID
	ActorRoles     []Role
	ActorBootstrap bool
	Action         string
	TargetType     string
	TargetID       *string
	Status         AuditStatus
	Summary        *string
	CreatedAt      time.Time
}
