package dto

type GetAccountByIdInput struct {
	ID string `path:"id" format:"uuid" doc:"Account ID"`
}
