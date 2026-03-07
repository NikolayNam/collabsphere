package dto

type CreateAccountInput struct {
	Body struct {
		Email     string `json:"email" required:"true" format:"email"`
		Password  string `json:"password" required:"true" minLength:"8"`
		FirstName string `json:"firstName" required:"true" minLength:"1" maxLength:"200"`
		LastName  string `json:"lastName" required:"true" minLength:"1" maxLength:"200"`
	}
}
