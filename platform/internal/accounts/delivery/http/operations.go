package http

import "github.com/danielgtaylor/huma/v2"

var createAccountOp = huma.Operation{
	OperationID: "create-account",
	Method:      "POST",
	Path:        "/accounts",
	Tags:        []string{"Accounts"},
	Summary:     "Create account",
}

var getAccountByIdOp = huma.Operation{
	OperationID: "get-account",
	Method:      "GET",
	Path:        "/accounts/{id}",
	Tags:        []string{"Accounts"},
	Summary:     "Get account by id",
}

var getAccountByEmailOp = huma.Operation{
	OperationID: "get-account-by-email",
	Method:      "GET",
	Path:        "/accounts/by-email",
	Tags:        []string{"Accounts"},
	Summary:     "Get account by email",
}

var getMyAccountOp = huma.Operation{
	OperationID: "get-my-account",
	Method:      "GET",
	Path:        "/accounts/me",
	Tags:        []string{"Accounts"},
	Summary:     "Get current account profile",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var updateMyAccountOp = huma.Operation{
	OperationID: "update-my-account",
	Method:      "PATCH",
	Path:        "/accounts/me",
	Tags:        []string{"Accounts"},
	Summary:     "Update current account profile",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createAvatarUploadOp = huma.Operation{
	OperationID: "create-account-avatar-upload",
	Method:      "POST",
	Path:        "/accounts/me/avatar-upload",
	Tags:        []string{"Accounts"},
	Summary:     "Create presigned upload for account avatar",
	Description: "This endpoint does not accept multipart file content. Send JSON metadata (fileName, contentType, sizeBytes, checksumSHA256) to receive a presigned upload URL. Then upload the raw file bytes with HTTP PUT to body.uploadUrl. After the upload succeeds, call PATCH /api/v1/accounts/me with avatarObjectId = body.objectId.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var uploadMyAvatarOp = huma.Operation{
	OperationID: "upload-my-account-avatar",
	Method:      "POST",
	Path:        "/accounts/me/avatar",
	Tags:        []string{"Accounts"},
	Summary:     "Upload account avatar directly",
	Description: "Single-step avatar upload using multipart/form-data. Send the image file in the `file` field. The backend uploads the object to S3-compatible storage and immediately attaches it to the current account profile.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
