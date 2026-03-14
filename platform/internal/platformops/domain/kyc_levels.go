package domain

import (
	"time"

	"github.com/google/uuid"
)

type KYCLevel struct {
	ID                    uuid.UUID
	Scope                 string
	Code                  string
	Name                  string
	Rank                  int
	IsActive              bool
	RequiredDocumentTypes []KYCLevelRequirement
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type KYCLevelRequirement struct {
	DocumentType string
	MinCount     int
}

type KYCLevelAssignment struct {
	LevelID   *uuid.UUID
	LevelCode *string
	LevelName *string
	IssuedAt  *time.Time
}
