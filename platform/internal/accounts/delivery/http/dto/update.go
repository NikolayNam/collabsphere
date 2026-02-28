package dto

type UpdateAccountByIDInput struct {
	ID   string `path:"id" doc:"User ID (UUID)"`
	Body struct {
		Email     *string `json:"email,omitempty"`
		FirstName *string `json:"first_name,omitempty"`
		LastName  *string `json:"last_name,omitempty"`
		Status    *string `json:"status,omitempty"`
	}
}

type UpdateAccountByEmailInput struct {
	Email string `query:"email" required:"true" doc:"Lookup email"`
	Body  struct {
		Email     *string `json:"email,omitempty"` // Новый email
		FirstName *string `json:"first_name,omitempty"`
		LastName  *string `json:"last_name,omitempty"`
		Status    *bool   `json:"is_active,omitempty"`
	}
}
