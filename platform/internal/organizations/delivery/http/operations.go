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

var getCooperationApplicationOp = huma.Operation{
	OperationID: "get-organization-cooperation-application",
	Method:      "GET",
	Path:        "/organizations/{id}/cooperation-application",
	Tags:        []string{"Organizations"},
	Summary:     "Get organization cooperation application",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var updateCooperationApplicationOp = huma.Operation{
	OperationID: "update-organization-cooperation-application",
	Method:      "PATCH",
	Path:        "/organizations/{id}/cooperation-application",
	Tags:        []string{"Organizations"},
	Summary:     "Create or update organization cooperation application",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var submitCooperationApplicationOp = huma.Operation{
	OperationID: "submit-organization-cooperation-application",
	Method:      "POST",
	Path:        "/organizations/{id}/cooperation-application/submit",
	Tags:        []string{"Organizations"},
	Summary:     "Submit organization cooperation application for review",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createCooperationPriceListUploadOp = huma.Operation{
	OperationID: "create-organization-price-list-upload",
	Method:      "POST",
	Path:        "/organizations/{id}/cooperation-application/price-list-upload",
	Tags:        []string{"Organizations"},
	Summary:     "Create presigned upload for cooperation price list",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createOrganizationLegalDocumentUploadOp = huma.Operation{
	OperationID: "create-organization-legal-document-upload",
	Method:      "POST",
	Path:        "/organizations/{id}/legal-documents/upload",
	Tags:        []string{"Organizations"},
	Summary:     "Create presigned upload for organization legal document",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createOrganizationLegalDocumentOp = huma.Operation{
	OperationID: "create-organization-legal-document",
	Method:      "POST",
	Path:        "/organizations/{id}/legal-documents",
	Tags:        []string{"Organizations"},
	Summary:     "Register uploaded organization legal document",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listOrganizationLegalDocumentsOp = huma.Operation{
	OperationID: "list-organization-legal-documents",
	Method:      "GET",
	Path:        "/organizations/{id}/legal-documents",
	Tags:        []string{"Organizations"},
	Summary:     "List organization legal documents",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getOrganizationLegalDocumentAnalysisOp = huma.Operation{
	OperationID: "get-organization-legal-document-analysis",
	Method:      "GET",
	Path:        "/organizations/{id}/legal-documents/{document_id}/analysis",
	Tags:        []string{"Organizations"},
	Summary:     "Get machine analysis result for organization legal document",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var reprocessOrganizationLegalDocumentAnalysisOp = huma.Operation{
	OperationID: "reprocess-organization-legal-document-analysis",
	Method:      "POST",
	Path:        "/organizations/{id}/legal-documents/{document_id}/analysis",
	Tags:        []string{"Organizations"},
	Summary:     "Requeue machine analysis for organization legal document",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
