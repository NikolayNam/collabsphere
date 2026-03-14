package dto

import (
	"time"

	"github.com/google/uuid"
)

type AccountResponse struct {
	Status int `json:"-"`
	Body   struct {
		ID             uuid.UUID   `json:"id"`
		Email          string      `json:"email"`
		ZitadelUserID  *string     `json:"zitadelUserId,omitempty"`
		DisplayName    *string     `json:"displayName,omitempty"`
		AvatarObjectID *uuid.UUID  `json:"avatarObjectId,omitempty"`
		VideoObjectIDs []uuid.UUID `json:"videoObjectIds,omitempty"`
		IsActive       bool        `json:"isActive"`
	}
}

type AccountProfileResponse struct {
	Status int `json:"-"`
	Body   struct {
		ID             uuid.UUID   `json:"id"`
		Email          string      `json:"email"`
		ZitadelUserID  *string     `json:"zitadelUserId,omitempty"`
		DisplayName    *string     `json:"displayName,omitempty"`
		AvatarObjectID *uuid.UUID  `json:"avatarObjectId,omitempty"`
		VideoObjectIDs []uuid.UUID `json:"videoObjectIds,omitempty"`
		Bio            *string     `json:"bio,omitempty"`
		Phone          *string     `json:"phone,omitempty"`
		Locale         *string     `json:"locale,omitempty"`
		Timezone       *string     `json:"timezone,omitempty"`
		Website        *string     `json:"website,omitempty"`
		IsActive       bool        `json:"isActive"`
		CreatedAt      time.Time   `json:"createdAt"`
		UpdatedAt      *time.Time  `json:"updatedAt,omitempty"`
	}
}

type AccountVideosResponse struct {
	Status int `json:"-"`
	Body   struct {
		Items []AccountVideoItem `json:"items"`
	}
}

type AccountVideoResponse struct {
	Status int              `json:"-"`
	Body   AccountVideoItem `json:"body"`
}

type AccountVideoItem struct {
	ID          uuid.UUID  `json:"id"`
	ObjectID    uuid.UUID  `json:"objectId"`
	FileName    string     `json:"fileName"`
	ContentType *string    `json:"contentType,omitempty"`
	SizeBytes   int64      `json:"sizeBytes"`
	CreatedAt   time.Time  `json:"createdAt"`
	SortOrder   int64      `json:"sortOrder"`
}
