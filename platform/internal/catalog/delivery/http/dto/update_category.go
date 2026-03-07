package dto

type UpdateProductCategoryInput struct {
	OrganizationID string `path:"organization_id"`
	CategoryID     string `path:"category_id"`
	Body           struct {
		ParentID  *string `json:"parentId,omitempty" format:"uuid"`
		Code      *string `json:"code,omitempty" maxLength:"128"`
		Name      *string `json:"name,omitempty" maxLength:"255"`
		SortOrder *int64  `json:"sortOrder,omitempty"`
	}
}
