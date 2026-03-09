package dto

import "github.com/danielgtaylor/huma/v2"

type UploadProductImportForm struct {
	File huma.FormFile `form:"file" required:"true" doc:"CSV import file with categories and products. Upload it directly with multipart/form-data."`
}

type UploadProductImportInput struct {
	OrganizationID string `path:"organization_id"`
	RawBody        huma.MultipartFormFiles[UploadProductImportForm]
}
