package dto

import (
	"time"

	"github.com/google/uuid"
)

type AccountResponse struct {
	Status int `json:"-"`
	Body   struct {
		ID             uuid.UUID  `json:"id"`
		Email          string     `json:"email"`
		DisplayName    *string    `json:"displayName,omitempty"`
		AvatarObjectID *uuid.UUID `json:"avatarObjectId,omitempty"`
		IsActive       bool       `json:"isActive"`
	}
}

type AccountProfileResponse struct {
	Status int `json:"-"`
	Body   struct {
		ID             uuid.UUID  `json:"id"`
		Email          string     `json:"email"`
		DisplayName    *string    `json:"displayName,omitempty"`
		AvatarObjectID *uuid.UUID `json:"avatarObjectId,omitempty"`
		Bio            *string    `json:"bio,omitempty"`
		Phone          *string    `json:"phone,omitempty"`
		Locale         *string    `json:"locale,omitempty"`
		Timezone       *string    `json:"timezone,omitempty"`
		Website        *string    `json:"website,omitempty"`
		IsActive       bool       `json:"isActive"`
		CreatedAt      time.Time  `json:"createdAt"`
		UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
	}
}

type UploadResponse struct {
	Status int `json:"-"`
	Body   struct {
		ObjectID     uuid.UUID `json:"objectId" doc:"Internal object ID. Use it in the next PATCH /accounts/me call as avatarObjectId after the file upload succeeds."`
		Bucket       string    `json:"bucket" doc:"Storage bucket where the file will be uploaded."`
		ObjectKey    string    `json:"objectKey" doc:"Storage object key reserved for this upload."`
		UploadMethod string    `json:"uploadMethod" doc:"HTTP method to use when uploading raw file bytes to uploadUrl. Usually PUT."`
		UploadURL    string    `json:"uploadUrl" doc:"Presigned storage URL. Send the raw file bytes to this URL, not JSON metadata."`
		ExpiresAt    time.Time `json:"expiresAt" doc:"Expiration time of the presigned upload URL."`
		FileName     string    `json:"fileName" doc:"Original file name stored in object metadata."`
		SizeBytes    int64     `json:"sizeBytes" doc:"Declared file size in bytes."`
	}
}
