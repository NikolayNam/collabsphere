package dto

import (
	"time"

	"github.com/google/uuid"
)

type ListOrganizationReviewsInput struct {
	Status            string `query:"status" doc:"Optional cooperation review status filter: draft, submitted, under_review, approved, rejected, needs_info."`
	OrganizationID    string `query:"organizationId" doc:"Optional organization filter."`
	ReviewerAccountID string `query:"reviewerAccountId" doc:"Optional reviewer account filter."`
	Q                 string `query:"q" doc:"Optional search across organization name, slug, company name, and confirmation email."`
	Limit             int    `query:"limit" doc:"Max items to return. Defaults to 50, capped at 200."`
	Offset            int    `query:"offset" doc:"Pagination offset. Defaults to 0."`
}

type GetOrganizationReviewInput struct {
	OrganizationID string `path:"organizationId" required:"true" doc:"Organization id whose review card should be loaded."`
}

type TransitionCooperationApplicationReviewInput struct {
	OrganizationID string `path:"organizationId" required:"true" doc:"Organization id whose cooperation application should transition."`
	Body           struct {
		TargetStatus string  `json:"targetStatus" required:"true" doc:"Target review status: under_review, approved, rejected, needs_info."`
		ReviewNote   *string `json:"reviewNote,omitempty" doc:"Optional reviewer note. Required for rejected and needs_info."`
	}
}

type TransitionLegalDocumentReviewInput struct {
	OrganizationID string `path:"organizationId" required:"true" doc:"Organization id whose legal document should transition."`
	DocumentID     string `path:"documentId" required:"true" doc:"Legal document id that should be reviewed."`
	Body           struct {
		TargetStatus string  `json:"targetStatus" required:"true" doc:"Target document review status: approved or rejected."`
		ReviewNote   *string `json:"reviewNote,omitempty" doc:"Optional reviewer note. Required for rejected."`
	}
}

type OrganizationReviewQueueResponse struct {
	Status int `json:"-"`
	Body   struct {
		Total int                           `json:"total"`
		Items []OrganizationReviewQueueItem `json:"items"`
	}
}

type OrganizationReviewQueueItem struct {
	OrganizationID           uuid.UUID  `json:"organizationId"`
	OrganizationName         string     `json:"organizationName"`
	OrganizationSlug         string     `json:"organizationSlug"`
	OrganizationIsActive     bool       `json:"organizationIsActive"`
	CooperationApplicationID uuid.UUID  `json:"cooperationApplicationId"`
	CooperationStatus        string     `json:"cooperationStatus"`
	CompanyName              *string    `json:"companyName,omitempty"`
	ConfirmationEmail        *string    `json:"confirmationEmail,omitempty"`
	ReviewerAccountID        *uuid.UUID `json:"reviewerAccountId,omitempty"`
	SubmittedAt              *time.Time `json:"submittedAt,omitempty"`
	ReviewedAt               *time.Time `json:"reviewedAt,omitempty"`
	CreatedAt                time.Time  `json:"createdAt"`
	UpdatedAt                *time.Time `json:"updatedAt,omitempty"`
}

type OrganizationReviewDetailResponse struct {
	Status int `json:"-"`
	Body   struct {
		Organization           OrganizationReviewOrganization            `json:"organization"`
		Domains                []OrganizationReviewDomain                `json:"domains"`
		CooperationApplication *OrganizationReviewCooperationApplication `json:"cooperationApplication,omitempty"`
		LegalDocuments         []OrganizationReviewLegalDocument         `json:"legalDocuments"`
		KYC                    *OrganizationReviewKYCRequirements        `json:"kyc,omitempty"`
	}
}

type CooperationApplicationReviewResponse struct {
	Status int `json:"-"`
	Body   OrganizationReviewCooperationApplication
}

type LegalDocumentReviewResponse struct {
	Status int `json:"-"`
	Body   OrganizationReviewLegalDocument
}

type OrganizationReviewOrganization struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	LogoObjectID *uuid.UUID `json:"logoObjectId,omitempty"`
	Description  *string    `json:"description,omitempty"`
	Website      *string    `json:"website,omitempty"`
	PrimaryEmail *string    `json:"primaryEmail,omitempty"`
	Phone        *string    `json:"phone,omitempty"`
	Address      *string    `json:"address,omitempty"`
	Industry     *string    `json:"industry,omitempty"`
	IsActive     bool       `json:"isActive"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type OrganizationReviewDomain struct {
	ID         uuid.UUID  `json:"id"`
	Hostname   string     `json:"hostname"`
	Kind       string     `json:"kind"`
	IsPrimary  bool       `json:"isPrimary"`
	IsVerified bool       `json:"isVerified"`
	VerifiedAt *time.Time `json:"verifiedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
}

type OrganizationReviewCooperationApplication struct {
	ID                    uuid.UUID  `json:"id"`
	OrganizationID        uuid.UUID  `json:"organizationId"`
	Status                string     `json:"status"`
	ConfirmationEmail     *string    `json:"confirmationEmail,omitempty"`
	CompanyName           *string    `json:"companyName,omitempty"`
	RepresentedCategories *string    `json:"representedCategories,omitempty"`
	MinimumOrderAmount    *string    `json:"minimumOrderAmount,omitempty"`
	DeliveryGeography     *string    `json:"deliveryGeography,omitempty"`
	SalesChannels         []string   `json:"salesChannels"`
	StorefrontURL         *string    `json:"storefrontURL,omitempty"`
	ContactFirstName      *string    `json:"contactFirstName,omitempty"`
	ContactLastName       *string    `json:"contactLastName,omitempty"`
	ContactJobTitle       *string    `json:"contactJobTitle,omitempty"`
	PriceListObjectID     *uuid.UUID `json:"priceListObjectId,omitempty"`
	ContactEmail          *string    `json:"contactEmail,omitempty"`
	ContactPhone          *string    `json:"contactPhone,omitempty"`
	PartnerCode           *string    `json:"partnerCode,omitempty"`
	ReviewNote            *string    `json:"reviewNote,omitempty"`
	ReviewerAccountID     *uuid.UUID `json:"reviewerAccountId,omitempty"`
	SubmittedAt           *time.Time `json:"submittedAt,omitempty"`
	ReviewedAt            *time.Time `json:"reviewedAt,omitempty"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             *time.Time `json:"updatedAt,omitempty"`
}

type OrganizationReviewLegalDocument struct {
	ID                  uuid.UUID                                    `json:"id"`
	OrganizationID      uuid.UUID                                    `json:"organizationId"`
	DocumentType        string                                       `json:"documentType"`
	Status              string                                       `json:"status"`
	ObjectID            uuid.UUID                                    `json:"objectId"`
	Title               string                                       `json:"title"`
	UploadedByAccountID *uuid.UUID                                   `json:"uploadedByAccountId,omitempty"`
	ReviewerAccountID   *uuid.UUID                                   `json:"reviewerAccountId,omitempty"`
	ReviewNote          *string                                      `json:"reviewNote,omitempty"`
	CreatedAt           time.Time                                    `json:"createdAt"`
	UpdatedAt           *time.Time                                   `json:"updatedAt,omitempty"`
	ReviewedAt          *time.Time                                   `json:"reviewedAt,omitempty"`
	Analysis            *OrganizationReviewLegalDocumentAnalysis     `json:"analysis,omitempty"`
	Verification        *OrganizationReviewLegalDocumentVerification `json:"verification,omitempty"`
}

type OrganizationReviewLegalDocumentAnalysis struct {
	ID                   uuid.UUID  `json:"id"`
	DocumentID           uuid.UUID  `json:"documentId"`
	OrganizationID       uuid.UUID  `json:"organizationId"`
	Status               string     `json:"status"`
	Provider             string     `json:"provider"`
	Summary              *string    `json:"summary,omitempty"`
	DetectedDocumentType *string    `json:"detectedDocumentType,omitempty"`
	ConfidenceScore      *float64   `json:"confidenceScore,omitempty"`
	RequestedAt          time.Time  `json:"requestedAt"`
	StartedAt            *time.Time `json:"startedAt,omitempty"`
	CompletedAt          *time.Time `json:"completedAt,omitempty"`
	UpdatedAt            *time.Time `json:"updatedAt,omitempty"`
	LastError            *string    `json:"lastError,omitempty"`
}

type OrganizationReviewLegalDocumentVerification struct {
	DocumentID           uuid.UUID                                          `json:"documentId"`
	OrganizationID       uuid.UUID                                          `json:"organizationId"`
	DocumentType         string                                             `json:"documentType"`
	DocumentStatus       string                                             `json:"documentStatus"`
	AnalysisStatus       *string                                            `json:"analysisStatus,omitempty"`
	Verdict              string                                             `json:"verdict"`
	Summary              string                                             `json:"summary"`
	DetectedDocumentType *string                                            `json:"detectedDocumentType,omitempty"`
	ConfidenceScore      *float64                                           `json:"confidenceScore,omitempty"`
	RequiredFields       []string                                           `json:"requiredFields"`
	MissingFields        []string                                           `json:"missingFields"`
	Issues               []OrganizationReviewLegalDocumentVerificationIssue `json:"issues"`
	CheckedAt            time.Time                                          `json:"checkedAt"`
}

type OrganizationReviewLegalDocumentVerificationIssue struct {
	Code     string  `json:"code"`
	Severity string  `json:"severity"`
	Message  string  `json:"message"`
	Field    *string `json:"field,omitempty"`
}

type OrganizationReviewKYCRequirements struct {
	OrganizationID      uuid.UUID                              `json:"organizationId"`
	Status              string                                 `json:"status"`
	DisabledReason      *string                                `json:"disabledReason,omitempty"`
	CurrentlyDue        []OrganizationReviewKYCRequirementItem `json:"currentlyDue"`
	PendingVerification []OrganizationReviewKYCRequirementItem `json:"pendingVerification"`
	EventuallyDue       []OrganizationReviewKYCRequirementItem `json:"eventuallyDue"`
	Errors              []OrganizationReviewKYCRequirementItem `json:"errors"`
	CheckedAt           time.Time                              `json:"checkedAt"`
}

type OrganizationReviewKYCRequirementItem struct {
	Code         string     `json:"code"`
	Category     string     `json:"category"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Field        *string    `json:"field,omitempty"`
	DocumentID   *uuid.UUID `json:"documentId,omitempty"`
	DocumentType *string    `json:"documentType,omitempty"`
	Reason       *string    `json:"reason,omitempty"`
}
