package dto

type CreateGroupInput struct {
	Body struct {
		Name        string  `json:"name" required:"true" maxLength:"255"`
		Slug        string  `json:"slug" required:"true" maxLength:"255"`
		Description *string `json:"description,omitempty" maxLength:"2000"`
	}
}
