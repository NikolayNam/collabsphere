package http

import "github.com/danielgtaylor/huma/v2"

var downloadMyAvatarOp = huma.Operation{
	OperationID: "download-my-account-avatar",
	Method:      "GET",
	Path:        "/accounts/me/avatar/download",
	Tags:        []string{"Accounts / Files"},
	Summary:     "Download current account avatar",
	Description: "Returns a short-lived download URL for the current account avatar.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var downloadOrganizationLogoOp = huma.Operation{
	OperationID: "download-organization-logo",
	Method:      "GET",
	Path:        "/organizations/{id}/logo/download",
	Tags:        []string{"Organizations / Files"},
	Summary:     "Download organization logo",
	Description: "Returns a short-lived download URL for the organization logo.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var downloadCooperationPriceListOp = huma.Operation{
	OperationID: "download-organization-price-list",
	Method:      "GET",
	Path:        "/organizations/{id}/cooperation-application/price-list/download",
	Tags:        []string{"Organizations / Files"},
	Summary:     "Download cooperation application price list",
	Description: "Returns a short-lived download URL for the organization cooperation application price list.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var downloadOrganizationLegalDocumentOp = huma.Operation{
	OperationID: "download-organization-legal-document",
	Method:      "GET",
	Path:        "/organizations/{id}/legal-documents/{document_id}/download",
	Tags:        []string{"Organizations / Files"},
	Summary:     "Download organization legal document",
	Description: "Returns a short-lived download URL for an organization legal document.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var downloadProductImportSourceOp = huma.Operation{
	OperationID: "download-product-import-source",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/product-imports/{batch_id}/source/download",
	Tags:        []string{"Organizations / Files"},
	Summary:     "Download product import source file",
	Description: "Returns a short-lived download URL for the CSV or source file linked to a product import batch.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var downloadChannelAttachmentOp = huma.Operation{
	OperationID: "download-channel-attachment",
	Method:      "GET",
	Path:        "/channels/{channel_id}/attachments/{object_id}/download",
	Tags:        []string{"Collab / Files"},
	Summary:     "Download channel attachment",
	Description: "Returns a short-lived download URL for a file already attached to a message in the channel.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listConferenceRecordingsOp = huma.Operation{
	OperationID: "list-conference-recordings",
	Method:      "GET",
	Path:        "/conferences/{conference_id}/recordings",
	Tags:        []string{"Collab / Files"},
	Summary:     "List conference recordings",
	Description: "Returns the recordings attached to a conference, ordered from newest to oldest.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var downloadConferenceRecordingOp = huma.Operation{
	OperationID: "download-conference-recording",
	Method:      "GET",
	Path:        "/conferences/{conference_id}/recordings/{recording_id}/download",
	Tags:        []string{"Collab / Files"},
	Summary:     "Download conference recording",
	Description: "Returns a short-lived download URL for a specific conference recording.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listMyFilesOp = huma.Operation{
	OperationID: "list-my-files",
	Method:      "GET",
	Path:        "/accounts/me/files",
	Tags:        []string{"Accounts / Files"},
	Summary:     "List files directly linked to the current account",
	Description: "Returns files explicitly attached to the authenticated account. In the current model this is limited to account-owned profile media such as the avatar.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listOrganizationFilesOp = huma.Operation{
	OperationID: "list-organization-files",
	Method:      "GET",
	Path:        "/organizations/{id}/files",
	Tags:        []string{"Organizations / Files"},
	Summary:     "List files linked to an organization",
	Description: "Returns organization-bound files visible to an active organization member: logo, cooperation price list, legal documents, product import sources, product images, and order documents.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
