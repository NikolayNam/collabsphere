package http

import "github.com/danielgtaylor/huma/v2"

var addMemberOp = huma.Operation{
	OperationID: "add-organization-member",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/members",
	Tags:        []string{"Organizations / Members"},
	Summary:     "Add an organization member",
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
	Summary:     "Update an organization member",
	Description: "Updates a member role or active state. The route protects invariants such as keeping at least one active owner in the organization.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var removeMemberOp = huma.Operation{
	OperationID: "remove-organization-member",
	Method:      "DELETE",
	Path:        "/organizations/{organization_id}/members/{membership_id}",
	Tags:        []string{"Organizations / Members"},
	Summary:     "Remove an organization member",
	Description: "Removes an account membership from the organization. This route will reject operations that would remove the last active owner.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createInvitationOp = huma.Operation{
	OperationID: "create-organization-invitation",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/invitations",
	Tags:        []string{"Organizations / Members"},
	Summary:     "Create an organization invitation",
	Description: "Creates a time-limited organization invitation token for an email address and target role. Only organization owners or admins can issue invitations.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listInvitationsOp = huma.Operation{
	OperationID: "list-organization-invitations",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/invitations",
	Tags:        []string{"Organizations / Members"},
	Summary:     "List organization invitations",
	Description: "Returns current and historical organization invitations with their lifecycle status.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var acceptInvitationOp = huma.Operation{
	OperationID: "accept-organization-invitation",
	Method:      "POST",
	Path:        "/invitations/{token}/accept",
	Tags:        []string{"Organizations / Members"},
	Summary:     "Accept an organization invitation",
	Description: "Accepts an organization invitation for the authenticated account. The invitation email must match the authenticated account email.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createAccessRequestOp = huma.Operation{
	OperationID: "create-organization-access-request",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/access-requests",
	Tags:        []string{"Organizations / Members"},
	Summary:     "Create an organization access request",
	Description: "Creates a self-service request to join an organization. The request is reviewed later by organization owners or admins.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listAccessRequestsOp = huma.Operation{
	OperationID: "list-organization-access-requests",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/access-requests",
	Tags:        []string{"Organizations / Members"},
	Summary:     "List organization access requests",
	Description: "Returns organization access requests. Only organization owners/admins can read and review this queue.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var approveAccessRequestOp = huma.Operation{
	OperationID: "approve-organization-access-request",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/access-requests/{request_id}/approve",
	Tags:        []string{"Organizations / Members"},
	Summary:     "Approve an organization access request",
	Description: "Approves a pending request and grants organization membership to the requester using the requested role.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listOrganizationRolesOp = huma.Operation{
	OperationID: "list-organization-roles",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/roles",
	Tags:        []string{"Organizations / Members"},
	Summary:     "List organization roles",
	Description: "Returns system and custom organization roles. Only owners and admins can manage roles.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
var createOrganizationRoleOp = huma.Operation{
	OperationID: "create-organization-role",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/roles",
	Tags:        []string{"Organizations / Members"},
	Summary:     "Create a custom organization role",
	Description: "Creates a custom role extending a system base role. Code must be unique within the organization.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
var updateOrganizationRoleOp = huma.Operation{
	OperationID: "update-organization-role",
	Method:      "PATCH",
	Path:        "/organizations/{organization_id}/roles/{role_id}",
	Tags:        []string{"Organizations / Members"},
	Summary:     "Update an organization role",
	Description: "Updates a custom role. System roles cannot be modified. Role code is immutable.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
var deleteOrganizationRoleOp = huma.Operation{
	OperationID: "delete-organization-role",
	Method:      "DELETE",
	Path:        "/organizations/{organization_id}/roles/{role_id}",
	Tags:        []string{"Organizations / Members"},
	Summary:     "Delete an organization role",
	Description: "Soft-deletes a custom role. Fails if the role is assigned to any members.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var rejectAccessRequestOp = huma.Operation{
	OperationID: "reject-organization-access-request",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/access-requests/{request_id}/reject",
	Tags:        []string{"Organizations / Members"},
	Summary:     "Reject an organization access request",
	Description: "Rejects a pending access request without granting membership.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
