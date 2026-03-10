package http

import "github.com/danielgtaylor/huma/v2"

var loginOp = huma.Operation{
	OperationID: "auth-login",
	Method:      "POST",
	Path:        "/auth/login",
	Tags:        []string{"Auth"},
	Summary:     "Create a session",
	Description: "Authenticates an account using email and password credentials and returns access and refresh tokens.",
}

var zitadelLoginOp = huma.Operation{
	OperationID: "auth-zitadel-login",
	Method:      "GET",
	Path:        "/auth/zitadel/login",
	Tags:        []string{"Auth"},
	Summary:     "Start ZITADEL login",
	Description: "Creates a short-lived OIDC state/nonce pair and returns the authorization URL for ZITADEL.",
}

var zitadelCallbackOp = huma.Operation{
	OperationID: "auth-zitadel-callback",
	Method:      "GET",
	Path:        "/auth/zitadel/callback",
	Tags:        []string{"Auth"},
	Summary:     "Complete ZITADEL login",
	Description: "Exchanges the OIDC authorization code, verifies the external identity, links or provisions a local account, and returns local access and refresh tokens.",
}

var refreshOp = huma.Operation{
	OperationID: "auth-refresh",
	Method:      "POST",
	Path:        "/auth/refresh",
	Tags:        []string{"Auth"},
	Summary:     "Refresh the current session",
	Description: "Exchanges a valid refresh token for a fresh access token and refresh token pair.",
}

var logoutOp = huma.Operation{
	OperationID: "auth-logout",
	Method:      "POST",
	Path:        "/auth/logout",
	Tags:        []string{"Auth"},
	Summary:     "Revoke the current session",
	Description: "Revokes the current refresh session so the client can no longer use the associated refresh token.",
}

var meOp = huma.Operation{
	OperationID: "auth-me",
	Method:      "GET",
	Path:        "/auth/me",
	Tags:        []string{"Auth"},
	Summary:     "Get the authenticated principal",
	Description: "Returns the authenticated principal and its current profile snapshot. For account principals this includes the current account profile fields.",
	Security: []map[string][]string{
		{"bearerAuth": {}},
	},
}
