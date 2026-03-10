package dto

import "github.com/danielgtaylor/huma/v2"

type UploadProductVideoForm struct {
	File huma.FormFile `form:"file" contentType:"video/*" required:"true" doc:"Product video file. Upload it directly with multipart/form-data."`
}

type UploadProductVideoInput struct {
	OrganizationID string `path:"organization_id" format:"uuid" doc:"Organization ID"`
	ProductID      string `path:"product_id" format:"uuid" doc:"Product ID"`
	RawBody        huma.MultipartFormFiles[UploadProductVideoForm]
}

type ListProductVideosInput struct {
	OrganizationID string `path:"organization_id" format:"uuid" doc:"Organization ID"`
	ProductID      string `path:"product_id" format:"uuid" doc:"Product ID"`
}
