package dto

import (
	"time"

	"github.com/google/uuid"
)

type DownloadObjectInput struct {
	ObjectID string `path:"object_id" format:"uuid" doc:"Storage object ID"`
}

type ListMyFilesInput struct{}

type ListOrganizationFilesInput struct {
	OrganizationID string `path:"id" format:"uuid" doc:"Organization ID"`
}

type FileItem struct {
	ObjectID       uuid.UUID  `json:"objectId"`
	OrganizationID *uuid.UUID `json:"organizationId,omitempty"`
	FileName       string     `json:"fileName"`
	ContentType    *string    `json:"contentType,omitempty"`
	SizeBytes      int64      `json:"sizeBytes"`
	CreatedAt      time.Time  `json:"createdAt"`
	SourceType     string     `json:"sourceType" doc:"Logical source of the file in the system, for example account_avatar, organization_logo, product_image or legal_document."`
	SourceID       *uuid.UUID `json:"sourceId,omitempty" doc:"Identifier of the entity that references the file."`
}

type ListFilesResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []FileItem `json:"items"`
	}
}

type DownloadObjectResponse struct {
	Status int `json:"-"`
	Body   struct {
		ObjectID       uuid.UUID  `json:"objectId"`
		OrganizationID *uuid.UUID `json:"organizationId,omitempty"`
		FileName       string     `json:"fileName"`
		ContentType    *string    `json:"contentType,omitempty"`
		SizeBytes      int64      `json:"sizeBytes"`
		DownloadURL    string     `json:"downloadUrl" doc:"Short-lived presigned URL for downloading the file bytes from S3-compatible storage."`
		ExpiresAt      time.Time  `json:"expiresAt"`
		CreatedAt      time.Time  `json:"createdAt"`
	}
}
