package dto

import (
	"time"

	"github.com/google/uuid"
)

type ListUploadsInput struct {
	Status             *string `query:"status" doc:"Optional status filter: pending, ready, failed."`
	Purpose            *string `query:"purpose" doc:"Optional purpose filter: organization_legal_document, product_import."`
	OrganizationID     *string `query:"organizationId" doc:"Optional organization filter."`
	CreatedByAccountID *string `query:"createdByAccountId" doc:"Optional creator account filter."`
	Limit              *int    `query:"limit" doc:"Max items to return. Defaults to 50, capped at 200."`
	Offset             *int    `query:"offset" doc:"Pagination offset. Defaults to 0."`
}

type UploadQueueResponse struct {
	Status int `json:"-"`
	Body   struct {
		Total int               `json:"total"`
		Items []UploadQueueItem `json:"items"`
	}
}

type UploadQueueItem struct {
	ID                 uuid.UUID  `json:"id"`
	OrganizationID     *uuid.UUID `json:"organizationId,omitempty"`
	CreatedByAccountID uuid.UUID  `json:"createdByAccountId"`
	Purpose            string     `json:"purpose"`
	Status             string     `json:"status"`
	FileName           string     `json:"fileName"`
	ContentType        *string    `json:"contentType,omitempty"`
	DeclaredSizeBytes  int64      `json:"declaredSizeBytes"`
	ErrorCode          *string    `json:"errorCode,omitempty"`
	ErrorMessage       *string    `json:"errorMessage,omitempty"`
	ResultKind         *string    `json:"resultKind,omitempty"`
	ResultID           *uuid.UUID `json:"resultId,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          *time.Time `json:"updatedAt,omitempty"`
}
