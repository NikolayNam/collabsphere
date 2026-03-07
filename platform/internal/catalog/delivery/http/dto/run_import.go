package dto

type RunProductImportInput struct {
	OrganizationID string `path:"organization_id"`
	Body           struct {
		SourceObjectID string  `json:"sourceObjectId" required:"true" format:"uuid"`
		Mode           *string `json:"mode,omitempty" enum:"upsert"`
	}
}
