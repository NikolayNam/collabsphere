package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OrganizationReviewQueueQuery struct {
	Status            *string
	OrganizationID    *uuid.UUID
	ReviewerAccountID *uuid.UUID
	Search            *string
	Limit             int
	Offset            int
}

type OrganizationReviewQueueItem struct {
	OrganizationID           uuid.UUID
	OrganizationName         string
	OrganizationSlug         string
	OrganizationIsActive     bool
	CooperationApplicationID uuid.UUID
	CooperationStatus        string
	CompanyName              *string
	ConfirmationEmail        *string
	ReviewerAccountID        *uuid.UUID
	SubmittedAt              *time.Time
	ReviewedAt               *time.Time
	CreatedAt                time.Time
	UpdatedAt                *time.Time
}

type OrganizationReviewDetail struct {
	Organization           OrganizationReviewOrganization
	Domains                []OrganizationReviewDomain
	CooperationApplication *OrganizationReviewCooperationApplication
	LegalDocuments         []OrganizationReviewLegalDocument
	KYC                    *OrganizationReviewKYCRequirements
}

type OrganizationReviewOrganization struct {
	ID           uuid.UUID
	Name         string
	Slug         string
	LogoObjectID *uuid.UUID
	Description  *string
	Website      *string
	PrimaryEmail *string
	Phone        *string
	Address      *string
	Industry     *string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type OrganizationReviewDomain struct {
	ID         uuid.UUID
	Hostname   string
	Kind       string
	IsPrimary  bool
	IsVerified bool
	VerifiedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

type OrganizationReviewCooperationApplication struct {
	ID                    uuid.UUID
	OrganizationID        uuid.UUID
	Status                string
	ConfirmationEmail     *string
	CompanyName           *string
	RepresentedCategories *string
	MinimumOrderAmount    *string
	DeliveryGeography     *string
	SalesChannels         []string
	StorefrontURL         *string
	ContactFirstName      *string
	ContactLastName       *string
	ContactJobTitle       *string
	PriceListObjectID     *uuid.UUID
	ContactEmail          *string
	ContactPhone          *string
	PartnerCode           *string
	ReviewNote            *string
	ReviewerAccountID     *uuid.UUID
	SubmittedAt           *time.Time
	ReviewedAt            *time.Time
	CreatedAt             time.Time
	UpdatedAt             *time.Time
}

type OrganizationReviewLegalDocument struct {
	ID                  uuid.UUID
	OrganizationID      uuid.UUID
	DocumentType        string
	Status              string
	ObjectID            uuid.UUID
	Title               string
	UploadedByAccountID *uuid.UUID
	ReviewerAccountID   *uuid.UUID
	ReviewNote          *string
	CreatedAt           time.Time
	UpdatedAt           *time.Time
	ReviewedAt          *time.Time
	Analysis            *OrganizationReviewLegalDocumentAnalysis
	Verification        *OrganizationReviewLegalDocumentVerification
}

type OrganizationReviewLegalDocumentAnalysis struct {
	ID                   uuid.UUID
	DocumentID           uuid.UUID
	OrganizationID       uuid.UUID
	Status               string
	Provider             string
	Summary              *string
	ExtractedFieldsJSON  json.RawMessage
	DetectedDocumentType *string
	ConfidenceScore      *float64
	RequestedAt          time.Time
	StartedAt            *time.Time
	CompletedAt          *time.Time
	UpdatedAt            *time.Time
	LastError            *string
}

type OrganizationReviewLegalDocumentVerification struct {
	DocumentID           uuid.UUID
	OrganizationID       uuid.UUID
	DocumentType         string
	DocumentStatus       string
	AnalysisStatus       *string
	Verdict              string
	Summary              string
	DetectedDocumentType *string
	ConfidenceScore      *float64
	RequiredFields       []string
	MissingFields        []string
	Issues               []OrganizationReviewLegalDocumentVerificationIssue
	CheckedAt            time.Time
}

type OrganizationReviewLegalDocumentVerificationIssue struct {
	Code     string
	Severity string
	Message  string
	Field    *string
}

type OrganizationReviewKYCRequirements struct {
	OrganizationID      uuid.UUID
	Status              string
	DisabledReason      *string
	CurrentlyDue        []OrganizationReviewKYCRequirementItem
	PendingVerification []OrganizationReviewKYCRequirementItem
	EventuallyDue       []OrganizationReviewKYCRequirementItem
	Errors              []OrganizationReviewKYCRequirementItem
	CheckedAt           time.Time
}

type OrganizationReviewKYCRequirementItem struct {
	Code         string
	Category     string
	Title        string
	Description  string
	Field        *string
	DocumentID   *uuid.UUID
	DocumentType *string
	Reason       *string
}

type CooperationApplicationReviewPatch struct {
	Status            string
	ReviewNote        *string
	ReviewerAccountID *uuid.UUID
	ReviewedAt        *time.Time
	UpdatedAt         time.Time
}

type LegalDocumentReviewPatch struct {
	Status            string
	ReviewNote        *string
	ReviewerAccountID *uuid.UUID
	ReviewedAt        *time.Time
	UpdatedAt         time.Time
}
