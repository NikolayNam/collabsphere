package dto

type GetGroupByIDInput struct {
	ID string `path:"id" required:"true" format:"uuid"`
}
