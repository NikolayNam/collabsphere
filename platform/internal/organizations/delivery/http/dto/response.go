package dto

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationBody struct {
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
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
}

type OrganizationResponse struct {
	Status int              `json:"-"`
	Body   OrganizationBody `json:"body"`
}

type UploadResponse struct {
	Status int `json:"-"`
	Body   struct {
		ObjectID     uuid.UUID `json:"objectId" doc:"Internal object ID. Use it in the next PATCH/POST call after the file upload succeeds."`
		Bucket       string    `json:"bucket" doc:"Storage bucket where the file will be uploaded."`
		ObjectKey    string    `json:"objectKey" doc:"Storage object key reserved for this upload."`
		UploadMethod string    `json:"uploadMethod" doc:"HTTP method to use when uploading raw file bytes to uploadUrl. Usually PUT."`
		UploadURL    string    `json:"uploadUrl" doc:"Presigned storage URL. Send the raw file bytes to this URL, not JSON metadata."`
		ExpiresAt    time.Time `json:"expiresAt" doc:"Expiration time of the presigned upload URL."`
		FileName     string    `json:"fileName" doc:"Original file name stored in object metadata."`
		SizeBytes    int64     `json:"sizeBytes" doc:"Declared file size in bytes."`
	}
}

type MemberBody struct {
	ID        uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"accountId"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"created_at"`
}
type MembersResponse struct {
	OrganizationID uuid.UUID  `json:"organizationId"`
	Status         int        `json:"-"`
	Body           MemberBody `json:"body"`
}

type MembersListBody struct {
	Data []MemberBody `json:"data"`
}

type MembersListResponse struct {
	Status int             `json:"-"`
	Body   MembersListBody `json:"-"`
}
