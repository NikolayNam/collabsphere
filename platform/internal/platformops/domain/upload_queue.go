package domain

import (
	"time"

	"github.com/google/uuid"
)

type UploadQueueQuery struct {
	Status             *string
	Purpose            *string
	OrganizationID     *uuid.UUID
	CreatedByAccountID *uuid.UUID
	Limit              int
	Offset             int
}

type UploadQueueItem struct {
	ID                 uuid.UUID
	OrganizationID     *uuid.UUID
	CreatedByAccountID uuid.UUID
	Purpose            string
	Status             string
	FileName           string
	ContentType        *string
	DeclaredSizeBytes  int64
	ErrorCode          *string
	ErrorMessage       *string
	ResultKind         *string
	ResultID           *uuid.UUID
	CreatedAt          time.Time
	UpdatedAt          *time.Time
}
