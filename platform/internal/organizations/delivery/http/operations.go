package http

import "github.com/danielgtaylor/huma/v2"

var createOrganizationOp = huma.Operation{
	OperationID: "create-organization",
	Method:      "POST",
	Path:        "/organizations",
	Tags:        []string{"Organizations"},
	Summary:     "Create an organization",
	Description: "Creates a new organization for the authenticated account, optionally binds initial hostnames, and automatically provisions the first owner membership.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listMyOrganizationsOp = huma.Operation{
	OperationID: "list-my-organizations",
	Method:      "GET",
	Path:        "/organizations/my",
	Tags:        []string{"Organizations"},
	Summary:     "List my organizations",
	Description: "Returns the organizations where the authenticated account currently has an active membership, including the current membership role for each item.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getOrganizationByIdOp = huma.Operation{
	OperationID: "get-organization",
	Method:      "GET",
	Path:        "/organizations/{id}",
	Tags:        []string{"Organizations"},
	Summary:     "Get an organization by ID",
	Description: "Returns the organization profile by id, including current branding, attached video identifiers, and configured hostnames.",
}

var resolveOrganizationByHostOp = huma.Operation{
	OperationID: "resolve-organization-by-host",
	Method:      "GET",
	Path:        "/organizations/resolve-by-host",
	Tags:        []string{"Organizations"},
	Summary:     "Resolve an organization by host",
	Description: "Resolves a verified active hostname to its organization profile. Useful for tenant-aware routing based on subdomain or custom domain.",
}

var listOrganizationsOp = huma.Operation{
	OperationID: "list-organizations",
	Method:      "GET",
	Path:        "/organizations/list",
	Tags:        []string{"Organizations"},
	Summary:     "List active organizations",
	Description: "Returns active organizations (id, name, slug). Used for access request forms and similar flows.",
}

var listPublicKYCDirectoryOp = huma.Operation{
	OperationID: "list-public-kyc-directory-organizations",
	Method:      "GET",
	Path:        "/organizations/public/kyc-directory",
	Tags:        []string{"Organizations"},
	Summary:     "List public verified organizations",
	Description: "Returns organizations eligible for public listing by a dedicated KYC level and additional publication requirements (verified founding document, legal name, and verified primary domain).",
}

var updateOrganizationOp = huma.Operation{
	OperationID: "update-organization",
	Method:      "PATCH",
	Path:        "/organizations/{id}",
	Tags:        []string{"Organizations"},
	Summary:     "Update an organization profile",
	Description: "Updates mutable organization profile fields such as title, slug, domains, description, contact fields, branding references, and other business metadata.",
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

var getCooperationApplicationOp = huma.Operation{OperationID: "get-organization-cooperation-application", Method: "GET", Path: "/organizations/{id}/cooperation-application", Tags: []string{"Organizations / Onboarding"}, Summary: "Get a cooperation application", Description: "Returns the current cooperation application draft or submitted application associated with the organization.", Security: []map[string][]string{{"bearerAuth": {}}}}
var updateCooperationApplicationOp = huma.Operation{OperationID: "update-organization-cooperation-application", Method: "PATCH", Path: "/organizations/{id}/cooperation-application", Tags: []string{"Organizations / Onboarding"}, Summary: "Upsert a cooperation application", Description: "Creates or updates the cooperation application payload for the organization, including commercial and contact information required for review.", Security: []map[string][]string{{"bearerAuth": {}}}}
var submitCooperationApplicationOp = huma.Operation{OperationID: "submit-organization-cooperation-application", Method: "POST", Path: "/organizations/{id}/cooperation-application/submit", Tags: []string{"Organizations / Onboarding"}, Summary: "Submit a cooperation application", Description: "Moves the organization cooperation application from draft to review state after required data and supporting files are present.", Security: []map[string][]string{{"bearerAuth": {}}}}
var uploadCooperationPriceListOp = huma.Operation{OperationID: "upload-organization-price-list", Method: "POST", Path: "/organizations/{id}/cooperation-application/price-list", Tags: []string{"Organizations / Files"}, Summary: "Upload a cooperation price list", Description: "Single-step cooperation price list upload using multipart/form-data. Send the file in the `file` field. The backend uploads the object to S3-compatible storage and immediately attaches it to the organization cooperation application.", Security: []map[string][]string{{"bearerAuth": {}}}}
var publishAllCatalogOp = huma.Operation{OperationID: "publish-organization-catalog", Method: "POST", Path: "/organizations/{id}/catalog/publish-all", Tags: []string{"Organizations / Catalog"}, Summary: "Publish all catalog", Description: "Publishes all verified categories, products, and price list in one action. Requires all existing items to be in verified or published state. Only organization owners, admins, or catalog managers can perform this action.", Security: []map[string][]string{{"bearerAuth": {}}}}
var uploadOrganizationLegalDocumentOp = huma.Operation{OperationID: "upload-organization-legal-document", Method: "POST", Path: "/organizations/{id}/legal-documents/file", Tags: []string{"Organizations / Files"}, Summary: "Upload a legal document", Description: "Single-step organization legal document upload using multipart/form-data. Send `documentType`, optional `title`, and the document file in the `file` field. The backend uploads the object to S3-compatible storage and immediately registers the legal document in the system.", Security: []map[string][]string{{"bearerAuth": {}}}}
var listOrganizationLegalDocumentsOp = huma.Operation{OperationID: "list-organization-legal-documents", Method: "GET", Path: "/organizations/{id}/legal-documents", Tags: []string{"Organizations / Onboarding"}, Summary: "List legal documents", Description: "Returns the legal documents registered for the organization together with their current review or analysis status.", Security: []map[string][]string{{"bearerAuth": {}}}}
var getOrganizationKYCRequirementsOp = huma.Operation{OperationID: "get-organization-kyc-requirements", Method: "GET", Path: "/organizations/{id}/kyc/requirements", Tags: []string{"Organizations / Onboarding"}, Summary: "Get organization KYC requirements", Description: "Returns the current internal KYC requirements snapshot for the organization, including items currently due, pending verification, and blocking errors.", Security: []map[string][]string{{"bearerAuth": {}}}}
var getOrganizationLegalDocumentAnalysisOp = huma.Operation{OperationID: "get-organization-legal-document-analysis", Method: "GET", Path: "/organizations/{id}/legal-documents/{document_id}/analysis", Tags: []string{"Organizations / Onboarding"}, Summary: "Get a legal document analysis", Description: "Returns the latest machine analysis result for a legal document, including extracted fields and analysis status where available.", Security: []map[string][]string{{"bearerAuth": {}}}}
var getOrganizationLegalDocumentVerificationOp = huma.Operation{OperationID: "get-organization-legal-document-verification", Method: "GET", Path: "/organizations/{id}/legal-documents/{document_id}/verification", Tags: []string{"Organizations / Onboarding"}, Summary: "Get a legal document verification result", Description: "Builds a verification verdict on top of the latest machine analysis for a legal document, including detected type checks, confidence threshold checks, and required extracted field checks.", Security: []map[string][]string{{"bearerAuth": {}}}}
var reprocessOrganizationLegalDocumentAnalysisOp = huma.Operation{OperationID: "reprocess-organization-legal-document-analysis", Method: "POST", Path: "/organizations/{id}/legal-documents/{document_id}/analysis", Tags: []string{"Organizations / Onboarding"}, Summary: "Retry legal document analysis", Description: "Places the legal document back into the analysis queue so OCR or document-analysis processing can be retried.", Security: []map[string][]string{{"bearerAuth": {}}}}

var createOrganizationLegalDocumentUploadOp = huma.Operation{OperationID: "create-organization-legal-document-upload", Method: "POST", Path: "/organizations/{id}/legal-documents/uploads", Tags: []string{"Organizations / Files"}, Summary: "Create a legal document upload session", Description: "Creates a tracked upload session for a legal document and returns a presigned upload URL for direct-to-storage upload.", Security: []map[string][]string{{"bearerAuth": {}}}}
var completeOrganizationLegalDocumentUploadOp = huma.Operation{OperationID: "complete-organization-legal-document-upload", Method: "POST", Path: "/organizations/{id}/legal-documents/uploads/{upload_id}/complete", Tags: []string{"Organizations / Files"}, Summary: "Finalize a legal document upload", Description: "Finalizes a previously created legal document upload session after the file has been uploaded to object storage and registers the legal document in the organization record.", Security: []map[string][]string{{"bearerAuth": {}}}}

var getOrganizationKYCProfileOp = huma.Operation{OperationID: "get-organization-kyc", Method: "GET", Path: "/organizations/{id}/kyc", Tags: []string{"Organizations / KYC"}, Summary: "Get organization KYC profile", Description: "Returns the organization KYC profile state and uploaded KYC documents.", Security: []map[string][]string{{"bearerAuth": {}}}}
var updateOrganizationKYCProfileOp = huma.Operation{OperationID: "update-organization-kyc", Method: "PATCH", Path: "/organizations/{id}/kyc", Tags: []string{"Organizations / KYC"}, Summary: "Update organization KYC profile", Description: "Updates self-service organization KYC profile fields and allows moving status between draft and submitted.", Security: []map[string][]string{{"bearerAuth": {}}}}
var createOrganizationKYCDocumentUploadOp = huma.Operation{OperationID: "create-organization-kyc-document-upload", Method: "POST", Path: "/organizations/{id}/kyc/documents/uploads", Tags: []string{"Organizations / KYC"}, Summary: "Create organization KYC document upload", Description: "Creates a tracked upload session for organization KYC document and returns a presigned upload URL.", Security: []map[string][]string{{"bearerAuth": {}}}}
var completeOrganizationKYCDocumentUploadOp = huma.Operation{OperationID: "complete-organization-kyc-document-upload", Method: "POST", Path: "/organizations/{id}/kyc/documents/uploads/{upload_id}/complete", Tags: []string{"Organizations / KYC"}, Summary: "Finalize organization KYC document upload", Description: "Finalizes previously created organization KYC document upload and registers it in KYC profile.", Security: []map[string][]string{{"bearerAuth": {}}}}
