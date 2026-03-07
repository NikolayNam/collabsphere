package dto

type AddMemberInput struct {
	Body struct {
		OrganizationID string `path:"organizationId" required:"true" format:"uuid"`
		AccountID      string `json:"accountId" required:"true" format:"uuid"`
		Kind           string `json:"kind" required:"false" enum:"owner,member"`
	}
}
