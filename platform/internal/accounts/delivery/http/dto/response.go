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
		ObjectID  uuid.UUID `json:"objectId"`
		Bucket    string    `json:"bucket"`
		ObjectKey string    `json:"objectKey"`
		UploadURL string    `json:"uploadUrl"`
		ExpiresAt time.Time `json:"expiresAt"`
		FileName  string    `json:"fileName"`
		SizeBytes int64     `json:"sizeBytes"`
	}
}
