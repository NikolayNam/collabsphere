package http

import "github.com/danielgtaylor/huma/v2"

var createTenantOp = huma.Operation{
	OperationID: "create-tenant",
	Method:      "POST",
	Path:        "/tenants",
	Tags:        []string{"Tenants"},
	Summary:     "Create tenant",
	Description: "Creates a new tenant in the dedicated tenant schema and assigns the authenticated account as owner.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listMyTenantsOp = huma.Operation{
	OperationID: "list-my-tenants",
	Method:      "GET",
	Path:        "/tenants/my",
	Tags:        []string{"Tenants"},
	Summary:     "List my tenants",
	Description: "Lists tenants where the authenticated account has membership.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getTenantOp = huma.Operation{
	OperationID: "get-tenant",
	Method:      "GET",
	Path:        "/tenants/{id}",
	Tags:        []string{"Tenants"},
	Summary:     "Get tenant by ID",
	Description: "Returns tenant details when the actor has tenant membership.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var addTenantMemberOp = huma.Operation{
	OperationID: "add-tenant-member",
	Method:      "POST",
	Path:        "/tenants/{tenant_id}/members",
	Tags:        []string{"Tenants"},
	Summary:     "Add tenant member",
	Description: "Adds an account to the tenant membership list. Only tenant owners/admins can manage members.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listTenantMembersOp = huma.Operation{
	OperationID: "list-tenant-members",
	Method:      "GET",
	Path:        "/tenants/{tenant_id}/members",
	Tags:        []string{"Tenants"},
	Summary:     "List tenant members",
	Description: "Returns account memberships in the tenant.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var addTenantOrganizationOp = huma.Operation{
	OperationID: "add-tenant-organization",
	Method:      "POST",
	Path:        "/tenants/{tenant_id}/organizations",
	Tags:        []string{"Tenants"},
	Summary:     "Attach organization to tenant",
	Description: "Links an existing organization to the tenant graph container.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listTenantOrganizationsOp = huma.Operation{
	OperationID: "list-tenant-organizations",
	Method:      "GET",
	Path:        "/tenants/{tenant_id}/organizations",
	Tags:        []string{"Tenants"},
	Summary:     "List tenant organizations",
	Description: "Returns organizations linked to the tenant.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
