package dto

type AddAccountMemberInput struct {
	GroupID string `path:"group_id" required:"true" format:"uuid"`
	Body    struct {
		AccountID string `json:"accountId" required:"true" format:"uuid"`
		Role      string `json:"role,omitempty" enum:"owner,member"`
	}
}
