package dto

import (
	"time"

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

type CreateCooperationPriceListUploadInput struct {
	ID   string `path:"id" format:"uuid" doc:"Organization ID"`
	Body struct {
		FileName       string  `json:"fileName" required:"true" maxLength:"512"`
		ContentType    *string `json:"contentType,omitempty" maxLength:"255"`
		SizeBytes      *int64  `json:"sizeBytes,omitempty" minimum:"0"`
		ChecksumSHA256 *string `json:"checksumSHA256,omitempty" maxLength:"64"`
	}
}

type CreateLegalDocumentUploadInput struct {
	ID   string `path:"id" format:"uuid" doc:"Organization ID"`
	Body struct {
		DocumentType   string  `json:"documentType" required:"true" maxLength:"64"`
		FileName       string  `json:"fileName" required:"true" maxLength:"512"`
		ContentType    *string `json:"contentType,omitempty" maxLength:"255"`
		SizeBytes      *int64  `json:"sizeBytes,omitempty" minimum:"0"`
		ChecksumSHA256 *string `json:"checksumSHA256,omitempty" maxLength:"64"`
	}
}

type CreateOrganizationLegalDocumentInput struct {
	ID   string `path:"id" format:"uuid" doc:"Organization ID"`
	Body struct {
		DocumentType string    `json:"documentType" required:"true" maxLength:"64"`
		ObjectID     uuid.UUID `json:"objectId" required:"true"`
		Title        string    `json:"title" required:"true" maxLength:"255"`
	}
}

type ListOrganizationLegalDocumentsInput struct {
	ID string `path:"id" format:"uuid" doc:"Organization ID"`
}
