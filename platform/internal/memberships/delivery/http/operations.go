package http

import "github.com/danielgtaylor/huma/v2"

var addMemberOp = huma.Operation{
	OperationID: "add-organization-member",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/members",
	Tags:        []string{"Organizations / Members"},
	Summary:     "Add member to organization",
	Description: "Adds an account to the organization with the requested role. The route enforces organization-level membership rules and ownership invariants.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listMembersOp = huma.Operation{
	OperationID: "list-organization-members",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/members",
	Tags:        []string{"Organizations / Members"},
	Summary:     "List organization members",
	Description: "Returns the current organization membership roster with roles and activation state.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var updateMemberOp = huma.Operation{
	OperationID: "update-organization-member",
	Method:      "PATCH",
	Path:        "/organizations/{organization_id}/members/{membership_id}",
	Tags:        []string{"Organizations / Members"},
	Summary:     "Update organization member role or status",
	Description: "Updates a member role or active state. The route protects invariants such as keeping at least one active owner in the organization.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var removeMemberOp = huma.Operation{
	OperationID: "remove-organization-member",
	Method:      "DELETE",
	Path:        "/organizations/{organization_id}/members/{membership_id}",
	Tags:        []string{"Organizations / Members"},
	Summary:     "Remove organization member",
	Description: "Removes an account membership from the organization. This route will reject operations that would remove the last active owner.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
