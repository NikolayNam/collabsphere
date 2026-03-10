package dto

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type GetMyAccountInput struct{}

type ListMyVideosInput struct{}

type UpdateMyAccountProfileInput struct {
	Body struct {
		DisplayName    *string    `json:"displayName,omitempty" maxLength:"255"`
		AvatarObjectID *uuid.UUID `json:"avatarObjectId,omitempty"`
		ClearAvatar    bool       `json:"clearAvatar,omitempty"`
		Bio            *string    `json:"bio,omitempty" maxLength:"4096"`
		Phone          *string    `json:"phone,omitempty" maxLength:"32"`
		Locale         *string    `json:"locale,omitempty" maxLength:"16"`
		Timezone       *string    `json:"timezone,omitempty" maxLength:"64"`
		Website        *string    `json:"website,omitempty" maxLength:"512"`
	}
}

type UploadMyAvatarForm struct {
	File huma.FormFile `form:"file" contentType:"image/*" required:"true" doc:"Avatar image file. Upload it directly with multipart/form-data."`
}

type UploadMyAvatarInput struct {
	RawBody huma.MultipartFormFiles[UploadMyAvatarForm]
}

type UploadMyVideoForm struct {
	File huma.FormFile `form:"file" contentType:"video/*" required:"true" doc:"Account video file. Upload it directly with multipart/form-data."`
}

type UploadMyVideoInput struct {
	RawBody huma.MultipartFormFiles[UploadMyVideoForm]
}
