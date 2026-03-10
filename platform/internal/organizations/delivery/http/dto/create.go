package dto

type OrganizationDomainInput struct {
	Hostname  string `json:"hostname" required:"true" example:"tenant1.collabsphere.ru" maxLength:"253"`
	Kind      string `json:"kind,omitempty" example:"subdomain" enum:"subdomain,custom_domain"`
	IsPrimary bool   `json:"isPrimary,omitempty"`
}

type CreateOrganizationInput struct {
	Body struct {
		Name    string                    `json:"name" required:"true" example:"Acme Foods" maxLength:"255"`
		Slug    string                    `json:"slug" required:"true" example:"acme-foods" maxLength:"255"`
		Domains []OrganizationDomainInput `json:"domains,omitempty"`
	}
}
