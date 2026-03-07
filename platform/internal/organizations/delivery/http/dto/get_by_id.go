package dto

type GetOrganizationByIdInput struct {
	ID string `path:"id" format:"uuid" doc:"Organization ID"`
}
