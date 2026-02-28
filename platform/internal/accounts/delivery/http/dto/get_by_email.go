package dto

type GetAccountByEmailInput struct {
	Email string `query:"email" required:"true" format:"email" doc:"Lookup email"`
}
