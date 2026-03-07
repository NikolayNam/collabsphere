package dto

type ListMembersInput struct {
    OrganizationID string `path:"organization_id" required:"true" format:"uuid"`
}
