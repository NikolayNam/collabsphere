package dto

type GetProductImportInput struct {
	OrganizationID string `path:"organization_id"`
	BatchID        string `path:"batch_id" format:"uuid"`
}
