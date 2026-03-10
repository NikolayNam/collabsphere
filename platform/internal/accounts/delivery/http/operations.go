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
