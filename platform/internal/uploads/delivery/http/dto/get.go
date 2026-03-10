package dto

type GetUploadInput struct {
	ID string `path:"id" format:"uuid" doc:"Upload session ID"`
}
