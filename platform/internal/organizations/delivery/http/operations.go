package http

import "github.com/danielgtaylor/huma/v2"

var createOrganizationOp = huma.Operation{
	OperationID: "create-organization",
	Method:      "POST",
	Path:        "/organizations",
	Tags:        []string{"Organizations"},
	Summary:     "Create an organization",
	Description: "Creates a new organization for the authenticated account and automatically provisions the first owner membership.",
	Security: []map[string][]string{{"bearerAuth": {}}},
}

var getOrganizationByIdOp = huma.Operation{
	OperationID: "get-organization",
	Method:      "GET",
	Path:        "/organizations/{id}",
	Tags:        []string{"Organizations"},
	Summary:     "Get an organization by ID",
	Description: "Returns the organization profile by id, including current branding, attached video identifiers, and profile fields visible to the caller.",
}

var updateOrganizationOp = huma.Operation{
	OperationID: "update-organization",
	Method:      "PATCH",
	Path:        "/organizations/{id}",
	Tags:        []string{"Organizations"},
	Summary:     "Update an organization profile",
	Description: "Updates mutable organization profile fields such as title, slug, description, contact fields, branding references, and other business metadata.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var uploadOrganizationLogoOp = huma.Operation{
	OperationID: "upload-organization-logo",
	Method:      "POST",
	Path:        "/organizations/{id}/logo",
	Tags:        []string{"Organizations / Files"},
	Summary:     "Upload an organization logo",
	Description: "Single-step organization logo upload using multipart/form-data. Send the image file in the `file` field. The backend uploads the object to S3-compatible storage and immediately attaches it to the organization profile.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var uploadOrganizationVideoOp = huma.Operation{
	OperationID: "upload-organization-video",
	Method:      "POST",
	Path:        "/organizations/{id}/videos",
	Tags:        []string{"Organizations / Files"},
	Summary:     "Upload an organization video",
	Description: "Single-step organization video upload using multipart/form-data. Send the video file in the `file` field. The backend uploads the object to S3-compatible storage and appends it to the organization video collection.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listOrganizationVideosOp = huma.Operation{
	OperationID: "list-organization-videos",
	Method:      "GET",
	Path:        "/organizations/{id}/videos",
	Tags:        []string{"Organizations / Files"},
	Summary:     "List organization videos",
	Description: "Returns the videos attached to the organization in display order.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getCooperationApplicationOp = huma.Operation{ OperationID: "get-organization-cooperation-application", Method: "GET", Path: "/organizations/{id}/cooperation-application", Tags: []string{"Organizations / Onboarding"}, Summary: "Get a cooperation application", Description: "Returns the current cooperation application draft or submitted application associated with the organization.", Security: []map[string][]string{{"bearerAuth": {}}}, }
var updateCooperationApplicationOp = huma.Operation{ OperationID: "update-organization-cooperation-application", Method: "PATCH", Path: "/organizations/{id}/cooperation-application", Tags: []string{"Organizations / Onboarding"}, Summary: "Upsert a cooperation application", Description: "Creates or updates the cooperation application payload for the organization, including commercial and contact information required for review.", Security: []map[string][]string{{"bearerAuth": {}}}, }
var submitCooperationApplicationOp = huma.Operation{ OperationID: "submit-organization-cooperation-application", Method: "POST", Path: "/organizations/{id}/cooperation-application/submit", Tags: []string{"Organizations / Onboarding"}, Summary: "Submit a cooperation application", Description: "Moves the organization cooperation application from draft to review state after required data and supporting files are present.", Security: []map[string][]string{{"bearerAuth": {}}}, }
var uploadCooperationPriceListOp = huma.Operation{ OperationID: "upload-organization-price-list", Method: "POST", Path: "/organizations/{id}/cooperation-application/price-list", Tags: []string{"Organizations / Files"}, Summary: "Upload a cooperation price list", Description: "Single-step cooperation price list upload using multipart/form-data. Send the file in the `file` field. The backend uploads the object to S3-compatible storage and immediately attaches it to the organization cooperation application.", Security: []map[string][]string{{"bearerAuth": {}}}, }
var uploadOrganizationLegalDocumentOp = huma.Operation{ OperationID: "upload-organization-legal-document", Method: "POST", Path: "/organizations/{id}/legal-documents/file", Tags: []string{"Organizations / Files"}, Summary: "Upload a legal document", Description: "Single-step organization legal document upload using multipart/form-data. Send `documentType`, optional `title`, and the document file in the `file` field. The backend uploads the object to S3-compatible storage and immediately registers the legal document in the system.", Security: []map[string][]string{{"bearerAuth": {}}}, }
var listOrganizationLegalDocumentsOp = huma.Operation{ OperationID: "list-organization-legal-documents", Method: "GET", Path: "/organizations/{id}/legal-documents", Tags: []string{"Organizations / Onboarding"}, Summary: "List legal documents", Description: "Returns the legal documents registered for the organization together with their current review or analysis status.", Security: []map[string][]string{{"bearerAuth": {}}}, }
var getOrganizationLegalDocumentAnalysisOp = huma.Operation{ OperationID: "get-organization-legal-document-analysis", Method: "GET", Path: "/organizations/{id}/legal-documents/{document_id}/analysis", Tags: []string{"Organizations / Onboarding"}, Summary: "Get a legal document analysis", Description: "Returns the latest machine analysis result for a legal document, including extracted fields and analysis status where available.", Security: []map[string][]string{{"bearerAuth": {}}}, }
var reprocessOrganizationLegalDocumentAnalysisOp = huma.Operation{ OperationID: "reprocess-organization-legal-document-analysis", Method: "POST", Path: "/organizations/{id}/legal-documents/{document_id}/analysis", Tags: []string{"Organizations / Onboarding"}, Summary: "Retry legal document analysis", Description: "Places the legal document back into the analysis queue so OCR or document-analysis processing can be retried.", Security: []map[string][]string{{"bearerAuth": {}}}, }
