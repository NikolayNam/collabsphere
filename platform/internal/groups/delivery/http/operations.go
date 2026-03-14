package http

import "github.com/danielgtaylor/huma/v2"

var createGroupOp = huma.Operation{
	OperationID: "create-group",
	Method:      "POST",
	Path:        "/groups",
	Tags:        []string{"Groups"},
	Summary:     "Create a group",
	Description: "Creates a collaboration group and automatically adds the authenticated account as its first owner.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listMyGroupsOp = huma.Operation{
	OperationID: "list-my-groups",
	Method:      "GET",
	Path:        "/groups/my",
	Tags:        []string{"Groups"},
	Summary:     "List my groups",
	Description: "Returns the collaboration groups accessible to the authenticated account, including direct and organization-derived membership paths.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getGroupByIDOp = huma.Operation{
	OperationID: "get-group",
	Method:      "GET",
	Path:        "/groups/{id}",
	Tags:        []string{"Groups"},
	Summary:     "Get a group by ID",
	Description: "Returns a group visible to the authenticated actor. Group metadata is the ACL container used by channels and collaboration features.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var addAccountMemberOp = huma.Operation{
	OperationID: "add-group-account-member",
	Method:      "POST",
	Path:        "/groups/{group_id}/accounts",
	Tags:        []string{"Groups"},
	Summary:     "Add an account member",
	Description: "Adds an account member to the group. Ownership and membership rules are enforced by the group access model.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var addOrganizationMemberOp = huma.Operation{
	OperationID: "add-group-organization-member",
	Method:      "POST",
	Path:        "/groups/{group_id}/organizations",
	Tags:        []string{"Groups"},
	Summary:     "Add an organization member",
	Description: "Adds an organization to the group so its active members inherit access to the group's channels and conferences.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listMembersOp = huma.Operation{
	OperationID: "list-group-members",
	Method:      "GET",
	Path:        "/groups/{group_id}/members",
	Tags:        []string{"Groups"},
	Summary:     "List group members",
	Description: "Returns both account members and organization members linked to the group.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
