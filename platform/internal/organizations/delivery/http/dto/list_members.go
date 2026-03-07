package dto

type ListMembersInput struct {
	OrganizationID string `path:"organizationId" required:"true" format:"uuid"`
}
