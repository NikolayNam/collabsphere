package dto

type CreateProductImportUploadInput struct {
	OrganizationID string `path:"organization_id" format:"uuid" doc:"Organization ID"`
	Body           struct {
		FileName       string  `json:"fileName" required:"true" maxLength:"512"`
		ContentType    *string `json:"contentType,omitempty" maxLength:"255"`
		SizeBytes      *int64  `json:"sizeBytes,omitempty" minimum:"0"`
		ChecksumSHA256 *string `json:"checksumSha256,omitempty" minLength:"64" maxLength:"64"`
	}
}

type CompleteProductImportUploadInput struct {
	OrganizationID string `path:"organization_id" format:"uuid" doc:"Organization ID"`
	UploadID       string `path:"upload_id" format:"uuid" doc:"Upload session ID"`
	Body           struct {
		Mode *string `json:"mode,omitempty" enum:"upsert"`
	}
}
