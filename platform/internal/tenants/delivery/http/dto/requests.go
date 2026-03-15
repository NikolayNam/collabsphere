package dto

type CreateTenantInput struct {
	Body struct {
		Name        string  `json:"name" required:"true" maxLength:"255"`
		Slug        string  `json:"slug" required:"true" maxLength:"255"`
		Description *string `json:"description,omitempty" maxLength:"2000"`
	}
}

type ListMyTenantsInput struct{}

type GetTenantInput struct {
	ID string `path:"id" format:"uuid"`
}

type AddTenantMemberInput struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		AccountID string `json:"accountId" required:"true" format:"uuid"`
		Role      string `json:"role,omitempty" enum:"owner,admin,member"`
	}
}

type ListTenantMembersInput struct {
	TenantID string `path:"tenant_id" format:"uuid"`
}

type AddTenantOrganizationInput struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		OrganizationID string `json:"organizationId" required:"true" format:"uuid"`
	}
}

type ListTenantOrganizationsInput struct {
	TenantID string `path:"tenant_id" format:"uuid"`
}
