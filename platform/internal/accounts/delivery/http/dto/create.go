package dto

type CreateAccountInput struct {
	Body struct {
		Email       string  `json:"email" required:"true" format:"email"`
		Password    string  `json:"password" required:"true" minLength:"8"`
		DisplayName *string `json:"displayName,omitempty" maxLength:"255"`
	}
}
