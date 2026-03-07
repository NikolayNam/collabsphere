package dto

type GetProductByIDInput struct {
	OrganizationID string `path:"organization_id"`
	ProductID      string `path:"product_id"`
}
