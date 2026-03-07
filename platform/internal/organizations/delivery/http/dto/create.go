package dto

type CreateOrganizationInput struct {
    Body struct {
        Name string `json:"name" required:"true" example:"Acme Foods" maxLength:"255"`
        Slug string `json:"slug" required:"true" example:"acme-foods" maxLength:"255"`
    }
}
