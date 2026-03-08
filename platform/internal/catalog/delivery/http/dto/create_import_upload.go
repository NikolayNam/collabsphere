package dto

type CreateProductImportUploadInput struct {
	OrganizationID string `path:"organization_id"`
	Body           struct {
		FileName       string  `json:"fileName" required:"true" maxLength:"512" doc:"Original import file name. This endpoint does not receive file bytes; it only prepares a presigned upload."`
		ContentType    *string `json:"contentType,omitempty" maxLength:"255" doc:"Optional MIME type that should be used when uploading the file to the presigned URL."`
		SizeBytes      *int64  `json:"sizeBytes,omitempty" minimum:"0" doc:"Optional file size in bytes for metadata and validation."`
		ChecksumSHA256 *string `json:"checksumSha256,omitempty" minLength:"64" maxLength:"64" doc:"Optional SHA-256 checksum of the file contents in hex format."`
	}
}
