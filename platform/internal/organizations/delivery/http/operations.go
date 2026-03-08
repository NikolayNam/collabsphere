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

var updateOrganizationOp = huma.Operation{
	OperationID: "update-organization",
	Method:      "PATCH",
	Path:        "/organizations/{id}",
	Tags:        []string{"Organizations"},
	Summary:     "Update organization profile",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createOrganizationLogoUploadOp = huma.Operation{
	OperationID: "create-organization-logo-upload",
	Method:      "POST",
	Path:        "/organizations/{id}/logo-upload",
	Tags:        []string{"Organizations"},
	Summary:     "Create presigned upload for organization logo",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
