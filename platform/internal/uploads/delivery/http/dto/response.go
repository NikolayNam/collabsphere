package dto

import (
	"time"

	"github.com/google/uuid"
)

type UploadBody struct {
	ID                 uuid.UUID      `json:"id"`
	OrganizationID     *uuid.UUID     `json:"organizationId,omitempty"`
	ObjectID           uuid.UUID      `json:"objectId"`
	CreatedByAccountID uuid.UUID      `json:"createdByAccountId"`
	Purpose            string         `json:"purpose"`
	Status             string         `json:"status"`
	Bucket             string         `json:"bucket"`
	ObjectKey          string         `json:"objectKey"`
	FileName           string         `json:"fileName"`
	ContentType        *string        `json:"contentType,omitempty"`
	DeclaredSizeBytes  int64          `json:"declaredSizeBytes"`
	ActualSizeBytes    *int64         `json:"actualSizeBytes,omitempty"`
	ChecksumSHA256     *string        `json:"checksumSha256,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	ErrorCode          *string        `json:"errorCode,omitempty"`
	ErrorMessage       *string        `json:"errorMessage,omitempty"`
	ResultKind         *string        `json:"resultKind,omitempty"`
	ResultID           *uuid.UUID     `json:"resultId,omitempty"`
	CompletedAt        *time.Time     `json:"completedAt,omitempty"`
	ExpiresAt          *time.Time     `json:"expiresAt,omitempty"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          *time.Time     `json:"updatedAt,omitempty"`
	UploadURL          *string        `json:"uploadUrl,omitempty"`
}

type UploadResponse struct {
	Status int        `json:"-"`
	Body   UploadBody `json:"body"`
}
