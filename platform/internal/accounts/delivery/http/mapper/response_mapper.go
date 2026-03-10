package mapper

import (
	"time"

	"github.com/google/uuid"

	"github.com/NikolayNam/collabsphere/internal/accounts/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
)

func ToAccountResponse(a *domain.Account, status int) *dto.AccountResponse {
	if a == nil {
		return nil
	}

	return &dto.AccountResponse{
		Status: status,
		Body: struct {
			ID             uuid.UUID   `json:"id"`
			Email          string      `json:"email"`
			DisplayName    *string     `json:"displayName,omitempty"`
			AvatarObjectID *uuid.UUID  `json:"avatarObjectId,omitempty"`
			VideoObjectIDs []uuid.UUID `json:"videoObjectIds,omitempty"`
			IsActive       bool        `json:"isActive"`
		}{
			ID:             a.ID().UUID(),
			Email:          a.Email().String(),
			DisplayName:    a.DisplayName(),
			AvatarObjectID: a.AvatarObjectID(),
			VideoObjectIDs: nil,
			IsActive:       a.IsActive(),
		},
	}
}

func ToAccountProfileResponse(a *domain.Account, status int) *dto.AccountProfileResponse {
	if a == nil {
		return nil
	}

	return &dto.AccountProfileResponse{
		Status: status,
		Body: struct {
			ID             uuid.UUID   `json:"id"`
			Email          string      `json:"email"`
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
		}{
			ID:             a.ID().UUID(),
			Email:          a.Email().String(),
			DisplayName:    a.DisplayName(),
			AvatarObjectID: a.AvatarObjectID(),
			VideoObjectIDs: nil,
			Bio:            a.Bio(),
			Phone:          a.Phone(),
			Locale:         a.Locale(),
			Timezone:       a.Timezone(),
			Website:        a.Website(),
			IsActive:       a.IsActive(),
			CreatedAt:      a.CreatedAt(),
			UpdatedAt:      a.UpdatedAt(),
		},
	}
}
