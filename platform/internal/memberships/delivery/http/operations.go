package http

import "github.com/danielgtaylor/huma/v2"

var addMemberOp = huma.Operation{
	OperationID: "add-organization-member",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/members",
	Tags:        []string{"Members", "Organizations"},
	Summary:     "Add member to organization",
}

var listMembersOp = huma.Operation{
	OperationID: "list-organization-members",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/members",
	Tags:        []string{"Members", "Organizations"},
	Summary:     "List organization members",
}
