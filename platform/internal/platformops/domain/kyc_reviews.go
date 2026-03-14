package domain

import (
	"time"

	"github.com/google/uuid"
)

type KYCReviewQuery struct {
	Scope  *string
	Status *string
	Limit  int
	Offset int
}

type KYCReviewItem struct {
	ReviewID     string
	Scope        string
	SubjectID    uuid.UUID
	Status       string
	KYCLevelCode *string
	KYCLevelName *string
	LegalName    *string
	CountryCode  *string
	SubmittedAt  *time.Time
	ReviewedAt   *time.Time
	UpdatedAt    time.Time
}

type KYCReviewDetail struct {
	ReviewID           string
	Scope              string
	SubjectID          uuid.UUID
	Status             string
	KYCLevelCode       *string
	KYCLevelName       *string
	LegalName          *string
	CountryCode        *string
	RegistrationNumber *string
	TaxID              *string
	DocumentNumber     *string
	ResidenceAddress   *string
	ReviewNote         *string
	ReviewerAccountID  *uuid.UUID
	SubmittedAt        *time.Time
	ReviewedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Documents          []KYCDocumentReviewItem
	Events             []KYCReviewEvent
}

type KYCDecisionPatch struct {
	Scope             string
	SubjectID         uuid.UUID
	Status            string
	ReviewNote        *string
	ReviewerAccountID uuid.UUID
	ReviewedAt        time.Time
	UpdatedAt         time.Time
}

type KYCDocumentReviewItem struct {
	ID                uuid.UUID
	ObjectID          uuid.UUID
	DocumentType      string
	Title             string
	Status            string
	ReviewNote        *string
	ReviewerAccountID *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         *time.Time
	ReviewedAt        *time.Time
}

type KYCDocumentDecisionPatch struct {
	Scope             string
	SubjectID         uuid.UUID
	DocumentID        uuid.UUID
	Status            string
	ReviewNote        *string
	ReviewerAccountID uuid.UUID
	ReviewedAt        time.Time
	UpdatedAt         time.Time
}

type KYCReviewEvent struct {
	ID                uuid.UUID
	Scope             string
	SubjectID         uuid.UUID
	Decision          string
	Reason            *string
	ReviewerAccountID uuid.UUID
	CreatedAt         time.Time
}
