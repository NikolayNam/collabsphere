package http

import "github.com/danielgtaylor/huma/v2"

var createGroupOp = huma.Operation{
	OperationID: "create-group",
	Method:      "POST",
	Path:        "/groups",
	Tags:        []string{"Groups"},
	Summary:     "Create group",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getGroupByIDOp = huma.Operation{
	OperationID: "get-group",
	Method:      "GET",
	Path:        "/groups/{id}",
	Tags:        []string{"Groups"},
	Summary:     "Get group by id",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var addAccountMemberOp = huma.Operation{
	OperationID: "add-group-account-member",
	Method:      "POST",
	Path:        "/groups/{group_id}/accounts",
	Tags:        []string{"Groups"},
	Summary:     "Add account to group",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var addOrganizationMemberOp = huma.Operation{
	OperationID: "add-group-organization-member",
	Method:      "POST",
	Path:        "/groups/{group_id}/organizations",
	Tags:        []string{"Groups"},
	Summary:     "Add organization to group",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listMembersOp = huma.Operation{
	OperationID: "list-group-members",
	Method:      "GET",
	Path:        "/groups/{group_id}/members",
	Tags:        []string{"Groups"},
	Summary:     "List group members",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
