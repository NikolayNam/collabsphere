package dto

import (
	"time"

	"github.com/google/uuid"
)

type TokenResponse struct {
	Status int `json:"-"`
	Body   struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		TokenType    string `json:"tokenType"`
		ExpiresIn    int64  `json:"expiresIn"`
	}
}

type MeResponse struct {
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

type EmptyResponse struct {
	Status int `json:"-"`
}
