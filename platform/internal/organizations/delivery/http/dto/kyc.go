package dto

import (
	"time"

	uploadsdto "github.com/NikolayNam/collabsphere/internal/uploads/delivery/http/dto"
	"github.com/google/uuid"
)

type GetOrganizationKYCInput struct {
	ID string `path:"id" required:"true"`
}

type UpdateOrganizationKYCInput struct {
	ID   string `path:"id" required:"true"`
	Body struct {
		Status             *string `json:"status,omitempty" doc:"KYC profile status: draft or submitted for self-service updates."`
		LegalName          *string `json:"legalName,omitempty" maxLength:"255"`
		CountryCode        *string `json:"countryCode,omitempty" maxLength:"8"`
		RegistrationNumber *string `json:"registrationNumber,omitempty" maxLength:"128"`
		TaxID              *string `json:"taxId,omitempty" maxLength:"128"`
	}
}

type CreateOrganizationKYCDocumentUploadInput struct {
	ID   string `path:"id" required:"true"`
	Body struct {
		DocumentType   string  `json:"documentType" required:"true" maxLength:"64"`
		Title          *string `json:"title,omitempty" maxLength:"255"`
		FileName       string  `json:"fileName" required:"true" maxLength:"255"`
		ContentType    *string `json:"contentType,omitempty" maxLength:"255"`
		SizeBytes      *int64  `json:"sizeBytes,omitempty"`
		ChecksumSHA256 *string `json:"checksumSha256,omitempty" maxLength:"64"`
	}
}

type CompleteOrganizationKYCDocumentUploadInput struct {
	ID       string `path:"id" required:"true"`
	UploadID string `path:"upload_id" required:"true"`
}

type OrganizationKYCResponse struct {
	Status int `json:"-"`
	Body   struct {
		OrganizationID     uuid.UUID                 `json:"organizationId"`
		Status             string                    `json:"status"`
		LegalName          *string                   `json:"legalName,omitempty"`
		CountryCode        *string                   `json:"countryCode,omitempty"`
		RegistrationNumber *string                   `json:"registrationNumber,omitempty"`
		TaxID              *string                   `json:"taxId,omitempty"`
		ReviewNote         *string                   `json:"reviewNote,omitempty"`
		ReviewerAccountID  *uuid.UUID                `json:"reviewerAccountId,omitempty"`
		SubmittedAt        *time.Time                `json:"submittedAt,omitempty"`
		ReviewedAt         *time.Time                `json:"reviewedAt,omitempty"`
		CreatedAt          time.Time                 `json:"createdAt"`
		UpdatedAt          time.Time                 `json:"updatedAt"`
		Documents          []OrganizationKYCDocument `json:"documents"`
	}
}

type OrganizationKYCDocumentResponse struct {
	Status int                     `json:"-"`
	Body   OrganizationKYCDocument `json:"body"`
}

type OrganizationKYCDocument struct {
	ID                uuid.UUID  `json:"id"`
	OrganizationID    uuid.UUID  `json:"organizationId"`
	ObjectID          uuid.UUID  `json:"objectId"`
	DocumentType      string     `json:"documentType"`
	Title             string     `json:"title"`
	Status            string     `json:"status"`
	ReviewNote        *string    `json:"reviewNote,omitempty"`
	ReviewerAccountID *uuid.UUID `json:"reviewerAccountId,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         *time.Time `json:"updatedAt,omitempty"`
	ReviewedAt        *time.Time `json:"reviewedAt,omitempty"`
}

type OrganizationKYCUploadResponse = uploadsdto.UploadResponse
