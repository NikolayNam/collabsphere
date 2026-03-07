package dto

type DeleteProductCategoryInput struct {
	OrganizationID string `path:"organization_id"`
	CategoryID     string `path:"category_id"`
}
