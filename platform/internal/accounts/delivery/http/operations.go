package http

import "github.com/danielgtaylor/huma/v2"

var createAccountOp = huma.Operation{
	OperationID: "create-account",
	Method:      "POST",
	Path:        "/accounts",
	Tags:        []string{"Accounts"},
	Summary:     "Create an account",
	Description: "Creates a new account with email and password credentials. This is the public signup entrypoint for a new user.",
}

var getAccountByIdOp = huma.Operation{
	OperationID: "get-account",
	Method:      "GET",
	Path:        "/accounts/{id}",
	Tags:        []string{"Accounts"},
	Summary:     "Get an account by ID",
	Description: "Returns an account profile by account id. Intended for internal or backoffice-style lookups rather than self-service profile editing.",
}

var getAccountByEmailOp = huma.Operation{
	OperationID: "get-account-by-email",
	Method:      "GET",
	Path:        "/accounts/by-email",
	Tags:        []string{"Accounts"},
	Summary:     "Find an account by email",
	Description: "Resolves an account profile by email address. Useful for administrative or integration flows that need identity lookup before relationship management.",
}

var getMyAccountOp = huma.Operation{
	OperationID: "get-my-account",
	Method:      "GET",
	Path:        "/accounts/me",
	Tags:        []string{"Accounts"},
	Summary:     "Get the current account",
	Description: "Returns the profile of the authenticated account, including current profile fields, avatar metadata, and attached account video identifiers.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var updateMyAccountOp = huma.Operation{
	OperationID: "update-my-account",
	Method:      "PATCH",
	Path:        "/accounts/me",
	Tags:        []string{"Accounts"},
	Summary:     "Update the current account",
	Description: "Updates mutable profile fields of the authenticated account. This route does not upload file bytes directly; avatar and account videos have dedicated upload endpoints.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var uploadMyAvatarOp = huma.Operation{
	OperationID: "upload-my-account-avatar",
	Method:      "POST",
	Path:        "/accounts/me/avatar",
	Tags:        []string{"Accounts / Files"},
	Summary:     "Upload an account avatar",
	Description: "Single-step avatar upload using multipart/form-data. Send the image file in the `file` field. The backend uploads the object to S3-compatible storage and immediately attaches it to the current account profile.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var uploadMyVideoOp = huma.Operation{
	OperationID: "upload-my-account-video",
	Method:      "POST",
	Path:        "/accounts/me/videos",
	Tags:        []string{"Accounts / Files"},
	Summary:     "Upload an account video",
	Description: "Single-step account video upload using multipart/form-data. Send the video file in the `file` field. The backend uploads the object to S3-compatible storage and appends it to the current account video collection.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listMyVideosOp = huma.Operation{
	OperationID: "list-my-account-videos",
	Method:      "GET",
	Path:        "/accounts/me/videos",
	Tags:        []string{"Accounts / Files"},
	Summary:     "List account videos",
	Description: "Returns the videos attached to the authenticated account in display order.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getMyKYCOp = huma.Operation{
	OperationID: "get-my-account-kyc",
	Method:      "GET",
	Path:        "/accounts/me/kyc",
	Tags:        []string{"Accounts / KYC"},
	Summary:     "Get account KYC profile",
	Description: "Returns the current account KYC profile state and uploaded KYC documents.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var updateMyKYCOp = huma.Operation{
	OperationID: "update-my-account-kyc",
	Method:      "PATCH",
	Path:        "/accounts/me/kyc",
	Tags:        []string{"Accounts / KYC"},
	Summary:     "Update account KYC profile",
	Description: "Updates self-service account KYC profile fields and allows moving status between draft and submitted.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createMyKYCDocumentUploadOp = huma.Operation{
	OperationID: "create-my-account-kyc-document-upload",
	Method:      "POST",
	Path:        "/accounts/me/kyc/documents/uploads",
	Tags:        []string{"Accounts / KYC"},
	Summary:     "Create account KYC document upload",
	Description: "Creates an upload session and returns a presigned URL for account KYC document upload.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var completeMyKYCDocumentUploadOp = huma.Operation{
	OperationID: "complete-my-account-kyc-document-upload",
	Method:      "POST",
	Path:        "/accounts/me/kyc/documents/uploads/{upload_id}/complete",
	Tags:        []string{"Accounts / KYC"},
	Summary:     "Finalize account KYC document upload",
	Description: "Finalizes a previously created account KYC document upload session and registers the document in KYC profile.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
