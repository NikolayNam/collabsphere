package dto

import (
	"encoding/json"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type CooperationApplicationBody struct {
	ID                    uuid.UUID  `json:"id"`
	OrganizationID        uuid.UUID  `json:"organizationId"`
	Status                string     `json:"status"`
	ConfirmationEmail     *string    `json:"confirmationEmail,omitempty"`
	CompanyName           *string    `json:"companyName,omitempty"`
	RepresentedCategories *string    `json:"representedCategories,omitempty"`
	MinimumOrderAmount    *string    `json:"minimumOrderAmount,omitempty"`
	DeliveryGeography     *string    `json:"deliveryGeography,omitempty"`
	SalesChannels         []string   `json:"salesChannels"`
	StorefrontURL         *string    `json:"storefrontUrl,omitempty"`
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

type CooperationApplicationResponse struct {
	Status int                        `json:"-"`
	Body   CooperationApplicationBody `json:"body"`
}

type OrganizationLegalDocumentBody struct {
	ID                  uuid.UUID  `json:"id"`
	OrganizationID      uuid.UUID  `json:"organizationId"`
	DocumentType        string     `json:"documentType"`
	Status              string     `json:"status"`
	ObjectID            uuid.UUID  `json:"objectId"`
	Title               string     `json:"title"`
	UploadedByAccountID *uuid.UUID `json:"uploadedByAccountId,omitempty"`
	ReviewerAccountID   *uuid.UUID `json:"reviewerAccountId,omitempty"`
	ReviewNote          *string    `json:"reviewNote,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           *time.Time `json:"updatedAt,omitempty"`
	ReviewedAt          *time.Time `json:"reviewedAt,omitempty"`
}

type OrganizationLegalDocumentResponse struct {
	Status int                           `json:"-"`
	Body   OrganizationLegalDocumentBody `json:"body"`
}

type OrganizationLegalDocumentsResponse struct {
	Status int `json:"-"`
	Body   struct {
		Data []OrganizationLegalDocumentBody `json:"data"`
	} `json:"body"`
}

type OrganizationLegalDocumentAnalysisBody struct {
	ID                   uuid.UUID       `json:"id"`
	DocumentID           uuid.UUID       `json:"documentId"`
	OrganizationID       uuid.UUID       `json:"organizationId"`
	Status               string          `json:"status"`
	Provider             string          `json:"provider"`
	ExtractedText        *string         `json:"extractedText,omitempty"`
	Summary              *string         `json:"summary,omitempty"`
	ExtractedFields      json.RawMessage `json:"extractedFields"`
	DetectedDocumentType *string         `json:"detectedDocumentType,omitempty"`
	ConfidenceScore      *float64        `json:"confidenceScore,omitempty"`
	RequestedAt          time.Time       `json:"requestedAt"`
	StartedAt            *time.Time      `json:"startedAt,omitempty"`
	CompletedAt          *time.Time      `json:"completedAt,omitempty"`
	UpdatedAt            *time.Time      `json:"updatedAt,omitempty"`
	LastError            *string         `json:"lastError,omitempty"`
}

type OrganizationLegalDocumentAnalysisResponse struct {
	Status int                                   `json:"-"`
	Body   OrganizationLegalDocumentAnalysisBody `json:"body"`
}

type OrganizationLegalDocumentVerificationIssueBody struct {
	Code     string  `json:"code"`
	Severity string  `json:"severity"`
	Message  string  `json:"message"`
	Field    *string `json:"field,omitempty"`
}

type OrganizationLegalDocumentVerificationBody struct {
	DocumentID           uuid.UUID                                        `json:"documentId"`
	OrganizationID       uuid.UUID                                        `json:"organizationId"`
	DocumentType         string                                           `json:"documentType"`
	DocumentStatus       string                                           `json:"documentStatus"`
	AnalysisStatus       *string                                          `json:"analysisStatus,omitempty"`
	Verdict              string                                           `json:"verdict"`
	Summary              string                                           `json:"summary"`
	DetectedDocumentType *string                                          `json:"detectedDocumentType,omitempty"`
	ConfidenceScore      *float64                                         `json:"confidenceScore,omitempty"`
	RequiredFields       []string                                         `json:"requiredFields"`
	MissingFields        []string                                         `json:"missingFields"`
	Issues               []OrganizationLegalDocumentVerificationIssueBody `json:"issues"`
	CheckedAt            time.Time                                        `json:"checkedAt"`
}

type OrganizationLegalDocumentVerificationResponse struct {
	Status int                                       `json:"-"`
	Body   OrganizationLegalDocumentVerificationBody `json:"body"`
}

type OrganizationKYCRequirementItemBody struct {
	Code         string     `json:"code"`
	Category     string     `json:"category"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Field        *string    `json:"field,omitempty"`
	DocumentID   *uuid.UUID `json:"documentId,omitempty"`
	DocumentType *string    `json:"documentType,omitempty"`
	Reason       *string    `json:"reason,omitempty"`
}

type OrganizationKYCRequirementsBody struct {
	OrganizationID      uuid.UUID                            `json:"organizationId"`
	Status              string                               `json:"status"`
	DisabledReason      *string                              `json:"disabledReason,omitempty"`
	CurrentlyDue        []OrganizationKYCRequirementItemBody `json:"currentlyDue"`
	PendingVerification []OrganizationKYCRequirementItemBody `json:"pendingVerification"`
	EventuallyDue       []OrganizationKYCRequirementItemBody `json:"eventuallyDue"`
	Errors              []OrganizationKYCRequirementItemBody `json:"errors"`
	CheckedAt           time.Time                            `json:"checkedAt"`
}

type OrganizationKYCRequirementsResponse struct {
	Status int                             `json:"-"`
	Body   OrganizationKYCRequirementsBody `json:"body"`
}

type GetCooperationApplicationInput struct {
	ID string `path:"id" format:"uuid" doc:"Organization ID"`
}

type UpdateCooperationApplicationInput struct {
	ID   string `path:"id" format:"uuid" doc:"Organization ID"`
	Body struct {
		ConfirmationEmail     *string    `json:"confirmationEmail,omitempty" maxLength:"320"`
		CompanyName           *string    `json:"companyName,omitempty" maxLength:"255"`
		RepresentedCategories *string    `json:"representedCategories,omitempty" maxLength:"4096"`
		MinimumOrderAmount    *string    `json:"minimumOrderAmount,omitempty" maxLength:"128"`
		DeliveryGeography     *string    `json:"deliveryGeography,omitempty" maxLength:"4096"`
		SalesChannels         []string   `json:"salesChannels,omitempty"`
		StorefrontURL         *string    `json:"storefrontUrl,omitempty" maxLength:"512"`
		ContactFirstName      *string    `json:"contactFirstName,omitempty" maxLength:"128"`
		ContactLastName       *string    `json:"contactLastName,omitempty" maxLength:"128"`
		ContactJobTitle       *string    `json:"contactJobTitle,omitempty" maxLength:"128"`
		PriceListObjectID     *uuid.UUID `json:"priceListObjectId,omitempty"`
		ClearPriceList        bool       `json:"clearPriceList,omitempty"`
		ContactEmail          *string    `json:"contactEmail,omitempty" maxLength:"320"`
		ContactPhone          *string    `json:"contactPhone,omitempty" maxLength:"32"`
		PartnerCode           *string    `json:"partnerCode,omitempty" maxLength:"128"`
	}
}

type SubmitCooperationApplicationInput struct {
	ID string `path:"id" format:"uuid" doc:"Organization ID"`
}

type UploadCooperationPriceListForm struct {
	File huma.FormFile `form:"file" required:"true" doc:"Price list file. Upload it directly with multipart/form-data."`
}

type UploadCooperationPriceListInput struct {
	ID      string `path:"id" format:"uuid" doc:"Organization ID"`
	RawBody huma.MultipartFormFiles[UploadCooperationPriceListForm]
}

type ListOrganizationLegalDocumentsInput struct {
	ID string `path:"id" format:"uuid" doc:"Organization ID"`
}

type GetOrganizationLegalDocumentAnalysisInput struct {
	ID         string `path:"id" format:"uuid" doc:"Organization ID"`
	DocumentID string `path:"document_id" format:"uuid" doc:"Legal document ID"`
}

type ReprocessOrganizationLegalDocumentAnalysisInput struct {
	ID         string `path:"id" format:"uuid" doc:"Organization ID"`
	DocumentID string `path:"document_id" format:"uuid" doc:"Legal document ID"`
}

type GetOrganizationLegalDocumentVerificationInput struct {
	ID         string `path:"id" format:"uuid" doc:"Organization ID"`
	DocumentID string `path:"document_id" format:"uuid" doc:"Legal document ID"`
}

type GetOrganizationKYCRequirementsInput struct {
	ID string `path:"id" format:"uuid" doc:"Organization ID"`
}

type UploadOrganizationLegalDocumentForm struct {
	DocumentType string        `form:"documentType" required:"true" doc:"Legal document type."`
	Title        string        `form:"title" doc:"Optional document title. If omitted, the original file name is used."`
	File         huma.FormFile `form:"file" required:"true" doc:"Legal document file. Upload it directly with multipart/form-data."`
}

type UploadOrganizationLegalDocumentInput struct {
	ID      string `path:"id" format:"uuid" doc:"Organization ID"`
	RawBody huma.MultipartFormFiles[UploadOrganizationLegalDocumentForm]
}
