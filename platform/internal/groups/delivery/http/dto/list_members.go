package dto

type ListMembersInput struct {
	GroupID string `path:"group_id" required:"true" format:"uuid"`
}
