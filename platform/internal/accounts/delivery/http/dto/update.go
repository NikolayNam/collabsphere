package dto

import "github.com/google/uuid"

type GetMyAccountInput struct{}

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

type CreateAvatarUploadInput struct {
	Body struct {
		FileName       string  `json:"fileName" required:"true" maxLength:"512" doc:"Original file name. This endpoint does not receive file bytes; it only prepares a presigned upload."`
		ContentType    *string `json:"contentType,omitempty" maxLength:"255" doc:"Optional MIME type that should be used when uploading the file to the presigned URL."`
		SizeBytes      *int64  `json:"sizeBytes,omitempty" minimum:"0" doc:"Optional file size in bytes for metadata and validation."`
		ChecksumSHA256 *string `json:"checksumSHA256,omitempty" maxLength:"64" doc:"Optional SHA-256 checksum of the file contents in hex format."`
	}
}
