package domain

import (
	"time"

	"github.com/google/uuid"
)

type Purpose string

type Status string

type ResultKind string

const (
	PurposeOrganizationLegalDocument Purpose = "organization_legal_document"
	PurposeProductImport             Purpose = "product_import"
)

const (
	StatusPending Status = "pending"
	StatusReady   Status = "ready"
	StatusFailed  Status = "failed"
)

const (
	ResultKindOrganizationLegalDocument ResultKind = "organization_legal_document"
	ResultKindProductImportBatch        ResultKind = "product_import_batch"
)

type Upload struct {
	ID                 uuid.UUID
	OrganizationID     *uuid.UUID
	ObjectID           uuid.UUID
	CreatedByAccountID uuid.UUID
	Purpose            Purpose
	Status             Status
	Bucket             string
	ObjectKey          string
	FileName           string
	ContentType        *string
	DeclaredSizeBytes  int64
	ActualSizeBytes    *int64
	ChecksumSHA256     *string
	Metadata           map[string]any
	ErrorCode          *string
	ErrorMessage       *string
	ResultKind         *ResultKind
	ResultID           *uuid.UUID
	CompletedAt        *time.Time
	ExpiresAt          *time.Time
	CreatedAt          time.Time
	UpdatedAt          *time.Time
}
