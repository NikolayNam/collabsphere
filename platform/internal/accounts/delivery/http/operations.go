package http

import "github.com/danielgtaylor/huma/v2"

var createAccountOp = huma.Operation{
	OperationID: "create-account",
	Method:      "POST",
	Path:        "/accounts",
	Tags:        []string{"Accounts"},
	Summary:     "Create account",
	Description: "Creates a new account with email and password credentials. This is the public signup entrypoint for a new user.",
}

var getAccountByIdOp = huma.Operation{
	OperationID: "get-account",
	Method:      "GET",
	Path:        "/accounts/{id}",
	Tags:        []string{"Accounts"},
	Summary:     "Get account by id",
	Description: "Returns an account profile by account id. Intended for internal or backoffice-style lookups rather than self-service profile editing.",
}

var getAccountByEmailOp = huma.Operation{
	OperationID: "get-account-by-email",
	Method:      "GET",
	Path:        "/accounts/by-email",
	Tags:        []string{"Accounts"},
	Summary:     "Get account by email",
	Description: "Resolves an account profile by email address. Useful for administrative or integration flows that need identity lookup before relationship management.",
}

var getMyAccountOp = huma.Operation{
	OperationID: "get-my-account",
	Method:      "GET",
	Path:        "/accounts/me",
	Tags:        []string{"Accounts"},
	Summary:     "Get current account profile",
	Description: "Returns the profile of the authenticated account, including current profile fields and avatar attachment metadata when present.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var updateMyAccountOp = huma.Operation{
	OperationID: "update-my-account",
	Method:      "PATCH",
	Path:        "/accounts/me",
	Tags:        []string{"Accounts"},
	Summary:     "Update current account profile",
	Description: "Updates mutable profile fields of the authenticated account. This route does not upload file bytes directly; avatar upload has its own subject-specific endpoint.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var uploadMyAvatarOp = huma.Operation{
	OperationID: "upload-my-account-avatar",
	Method:      "POST",
	Path:        "/accounts/me/avatar",
	Tags:        []string{"Accounts / Files"},
	Summary:     "Upload account avatar directly",
	Description: "Single-step avatar upload using multipart/form-data. Send the image file in the `file` field. The backend uploads the object to S3-compatible storage and immediately attaches it to the current account profile.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
