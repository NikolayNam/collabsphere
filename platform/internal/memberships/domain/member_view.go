package domain

import (
    "time"

    "github.com/google/uuid"
)

type MemberView struct {
    MembershipID   uuid.UUID
    OrganizationID uuid.UUID
    AccountID      uuid.UUID
    Role           string
    IsActive       bool
    CreatedAt      time.Time
}
