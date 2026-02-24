package http

// CreateUserInput POST /users (в рамках X-Organization-ID)
type CreateUserInput struct {
	Body struct {
		Email     string `json:"email" required:"true" format:"email"`
		Password  string `json:"password" required:"true" minLength:"6"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Phone     string `json:"phone,omitempty"`
	}
}

// UpdateUserInput PUT /users/{user_id} (в рамках X-Organization-ID)
type UpdateUserInput struct {
	UserID uint `path:"user_id" doc:"User ID"`

	Body struct {
		FirstName *string `json:"first_name,omitempty"`
		LastName  *string `json:"last_name,omitempty"`
		Phone     *string `json:"phone,omitempty"`
		IsActive  *bool   `json:"is_active,omitempty" doc:"Deactivate membership in this organization"`
	}
}

type UserResponse struct {
	Body struct {
		ID        uint   `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`

		IsActive bool `json:"is_active"`
	}
}
