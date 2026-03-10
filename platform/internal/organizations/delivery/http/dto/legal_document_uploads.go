package dto

import "github.com/google/uuid"

type CreateOrganizationLegalDocumentUploadInput struct {
	ID   string `path:"id" format:"uuid" doc:"Organization ID"`
	Body struct {
		DocumentType   string  `json:"documentType" required:"true" maxLength:"64"`
		Title          *string `json:"title,omitempty" maxLength:"255"`
		FileName       string  `json:"fileName" required:"true" maxLength:"512"`
		ContentType    *string `json:"contentType,omitempty" maxLength:"255"`
		SizeBytes      *int64  `json:"sizeBytes,omitempty" minimum:"0"`
		ChecksumSHA256 *string `json:"checksumSha256,omitempty" minLength:"64" maxLength:"64"`
	}
}

type CompleteOrganizationLegalDocumentUploadInput struct {
	ID       string `path:"id" format:"uuid" doc:"Organization ID"`
	UploadID string `path:"upload_id" format:"uuid" doc:"Upload session ID"`
}

type CreateOrganizationLegalDocumentUploadResponse struct {
	Status int `json:"-"`
	Body   struct {
		UploadID uuid.UUID `json:"uploadId"`
	}
}
