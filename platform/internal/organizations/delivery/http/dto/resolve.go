package dto

type ResolveOrganizationByHostInput struct {
	Host string `query:"host" required:"true" example:"tenant1.collabsphere.ru" doc:"Organization hostname or URL to resolve"`
}
