package http

import "github.com/danielgtaylor/huma/v2"

var addMemberOp = huma.Operation{
	OperationID: "add-organization-member",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/members",
	Tags:        []string{"Members", "Organizations"},
	Summary:     "Add member to organization",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listMembersOp = huma.Operation{
	OperationID: "list-organization-members",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/members",
	Tags:        []string{"Members", "Organizations"},
	Summary:     "List organization members",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var updateMemberOp = huma.Operation{
	OperationID: "update-organization-member",
	Method:      "PATCH",
	Path:        "/organizations/{organization_id}/members/{membership_id}",
	Tags:        []string{"Members", "Organizations"},
	Summary:     "Update organization member role or status",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var removeMemberOp = huma.Operation{
	OperationID: "remove-organization-member",
	Method:      "DELETE",
	Path:        "/organizations/{organization_id}/members/{membership_id}",
	Tags:        []string{"Members", "Organizations"},
	Summary:     "Remove organization member",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
