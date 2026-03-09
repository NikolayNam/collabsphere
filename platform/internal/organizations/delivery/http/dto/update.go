package dto

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

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

type UploadOrganizationLogoForm struct {
	File huma.FormFile `form:"file" contentType:"image/*" required:"true" doc:"Organization logo image file. Upload it directly with multipart/form-data."`
}

type UploadOrganizationLogoInput struct {
	ID      string `path:"id" format:"uuid" doc:"Organization ID"`
	RawBody huma.MultipartFormFiles[UploadOrganizationLogoForm]
}
