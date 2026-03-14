package dto

import (
	"time"

	uploadsdto "github.com/NikolayNam/collabsphere/internal/uploads/delivery/http/dto"
	"github.com/google/uuid"
)

type GetMyKYCInput struct{}

type UpdateMyKYCInput struct {
	Body struct {
		Status           *string `json:"status,omitempty" doc:"KYC profile status: draft or submitted for self-service updates."`
		LegalName        *string `json:"legalName,omitempty" maxLength:"255"`
		CountryCode      *string `json:"countryCode,omitempty" maxLength:"8"`
		DocumentNumber   *string `json:"documentNumber,omitempty" maxLength:"128"`
		ResidenceAddress *string `json:"residenceAddress,omitempty" maxLength:"2048"`
	}
}

type CreateMyKYCDocumentUploadInput struct {
	Body struct {
		DocumentType   string  `json:"documentType" required:"true" maxLength:"64"`
		Title          *string `json:"title,omitempty" maxLength:"255"`
		FileName       string  `json:"fileName" required:"true" maxLength:"255"`
		ContentType    *string `json:"contentType,omitempty" maxLength:"255"`
		SizeBytes      *int64  `json:"sizeBytes,omitempty"`
		ChecksumSHA256 *string `json:"checksumSha256,omitempty" maxLength:"64"`
	}
}

type CompleteMyKYCDocumentUploadInput struct {
	UploadID string `path:"upload_id" required:"true" doc:"Upload id returned from create-my-account-kyc-document-upload."`
}

type AccountKYCResponse struct {
	Status int `json:"-"`
	Body   struct {
		AccountID        uuid.UUID            `json:"accountId"`
		Status           string               `json:"status"`
		LegalName        *string              `json:"legalName,omitempty"`
		CountryCode      *string              `json:"countryCode,omitempty"`
		DocumentNumber   *string              `json:"documentNumber,omitempty"`
		ResidenceAddress *string              `json:"residenceAddress,omitempty"`
		ReviewNote       *string              `json:"reviewNote,omitempty"`
		ReviewerAccount  *uuid.UUID           `json:"reviewerAccountId,omitempty"`
		SubmittedAt      *time.Time           `json:"submittedAt,omitempty"`
		ReviewedAt       *time.Time           `json:"reviewedAt,omitempty"`
		CreatedAt        time.Time            `json:"createdAt"`
		UpdatedAt        time.Time            `json:"updatedAt"`
		Documents        []AccountKYCDocument `json:"documents"`
	}
}

type AccountKYCDocumentResponse struct {
	Status int                `json:"-"`
	Body   AccountKYCDocument `json:"body"`
}

type AccountKYCDocument struct {
	ID              uuid.UUID  `json:"id"`
	AccountID       uuid.UUID  `json:"accountId"`
	ObjectID        uuid.UUID  `json:"objectId"`
	DocumentType    string     `json:"documentType"`
	Title           string     `json:"title"`
	Status          string     `json:"status"`
	ReviewNote      *string    `json:"reviewNote,omitempty"`
	ReviewerAccount *uuid.UUID `json:"reviewerAccountId,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       *time.Time `json:"updatedAt,omitempty"`
	ReviewedAt      *time.Time `json:"reviewedAt,omitempty"`
}

type AccountKYCUploadResponse = uploadsdto.UploadResponse
