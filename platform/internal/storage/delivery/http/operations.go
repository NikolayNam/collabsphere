package http

import "github.com/danielgtaylor/huma/v2"

var downloadObjectOp = huma.Operation{
	OperationID: "create-storage-object-download",
	Method:      "GET",
	Path:        "/storage/objects/{object_id}/download",
	Tags:        []string{"Storage"},
	Summary:     "Create presigned download for stored object",
	Description: "Returns a short-lived download URL for a file already stored in the system. Access is checked against the authenticated actor: own account avatar, organization membership for organization-bound files, and collab channel access for chat attachments and conference recordings.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listMyFilesOp = huma.Operation{
	OperationID: "list-my-files",
	Method:      "GET",
	Path:        "/accounts/me/files",
	Tags:        []string{"Storage"},
	Summary:     "List files directly linked to the current account",
	Description: "Returns files explicitly attached to the authenticated account. In the current model this is limited to account-owned profile media such as the avatar.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listOrganizationFilesOp = huma.Operation{
	OperationID: "list-organization-files",
	Method:      "GET",
	Path:        "/organizations/{id}/files",
	Tags:        []string{"Storage"},
	Summary:     "List files linked to an organization",
	Description: "Returns organization-bound files visible to an active organization member: logo, cooperation price list, legal documents, product import sources, product images, and order documents.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
