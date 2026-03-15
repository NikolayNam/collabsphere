package dto

type CreateProductCategoryInput struct {
	OrganizationID string `path:"organization_id"`
	Body           struct {
		ParentID  *string `json:"parentId,omitempty" format:"uuid"`
		Status    *string `json:"status,omitempty" maxLength:"24"`
		Code      string  `json:"code" required:"true" maxLength:"128"`
		Name      string  `json:"name" required:"true" maxLength:"255"`
		SortOrder int64   `json:"sortOrder"`
	}
}
