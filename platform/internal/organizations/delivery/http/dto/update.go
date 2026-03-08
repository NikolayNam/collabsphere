package dto

import "github.com/google/uuid"

type UpdateOrganizationInput struct {
	ID   string `path:"id" format:"uuid" doc:"Organization ID"`
	Body struct {
		Name         *string    `json:"name,omitempty" maxLength:"255"`
		Slug         *string    `json:"slug,omitempty" maxLength:"255"`
		LogoObjectID *uuid.UUID `json:"logoObjectId,omitempty"`
		ClearLogo    bool       `json:"clearLogo,omitempty"`
		Description  *string    `json:"description,omitempty" maxLength:"4096"`
		Website      *string    `json:"website,omitempty" maxLength:"512"`
		PrimaryEmail *string    `json:"primaryEmail,omitempty" maxLength:"320"`
		Phone        *string    `json:"phone,omitempty" maxLength:"32"`
		Address      *string    `json:"address,omitempty" maxLength:"4096"`
		Industry     *string    `json:"industry,omitempty" maxLength:"128"`
	}
}

type CreateOrganizationLogoUploadInput struct {
	ID   string `path:"id" format:"uuid" doc:"Organization ID"`
	Body struct {
		FileName       string  `json:"fileName" required:"true" maxLength:"512"`
		ContentType    *string `json:"contentType,omitempty" maxLength:"255"`
		SizeBytes      *int64  `json:"sizeBytes,omitempty" minimum:"0"`
		ChecksumSHA256 *string `json:"checksumSHA256,omitempty" maxLength:"64"`
	}
}
