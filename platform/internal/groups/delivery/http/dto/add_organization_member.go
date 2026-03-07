package dto

type AddOrganizationMemberInput struct {
	GroupID string `path:"group_id" required:"true" format:"uuid"`
	Body    struct {
		OrganizationID string `json:"organizationId" required:"true" format:"uuid"`
	}
}
