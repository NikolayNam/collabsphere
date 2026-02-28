package dto

type CreateAccountInput struct {
	Body struct {
		Email     string `json:"email" required:"true" format:"email"`
		Password  string `json:"password" required:"true" minLength:"6"`
		FirstName string `json:"first_name" required:"true" minLength:"1" maxLength:"200"`
		LastName  string `json:"last_name" required:"true" minLength:"1" maxLength:"200"`
	}
}
