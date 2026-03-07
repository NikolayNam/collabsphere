package http

import "github.com/danielgtaylor/huma/v2"

var createOrganizationOp = huma.Operation{
	OperationID: "create-organization",
	Method:      "POST",
	Path:        "/organizations",
	Tags:        []string{"Organizations"},
	Summary:     "Create organization",
	Security: []map[string][]string{
		{"bearerAuth": {}},
	},
}

var getOrganizationByIdOp = huma.Operation{
	OperationID: "get-organization",
	Method:      "GET",
	Path:        "/organizations/{id}",
	Tags:        []string{"Organizations"},
	Summary:     "Get organization by id",
}
