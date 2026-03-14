package http

import "github.com/danielgtaylor/huma/v2"

var loginOp = huma.Operation{
	OperationID: "auth-login",
	Method:      "POST",
	Path:        "/auth/login",
	Tags:        []string{"Auth"},
	Summary:     "Create a legacy password session",
	Description: "Authenticates an account using legacy email/password credentials and returns local access and refresh tokens. This route can be disabled by configuration when ZITADEL browser login is the primary path.",
}

var zitadelLoginOp = huma.Operation{
	OperationID: "auth-zitadel-login",
	Method:      "GET",
	Path:        "/auth/zitadel/login",
	Tags:        []string{"Auth"},
	Summary:     "Start ZITADEL login",
	Description: "Starts the browser-based ZITADEL login flow and responds with `303 See Other` to the self-hosted login UI on the login origin. Use the optional `return_to` query parameter to control where the callback should redirect after authentication.",
}

var zitadelSignupOp = huma.Operation{
	OperationID: "auth-zitadel-signup",
	Method:      "GET",
	Path:        "/auth/zitadel/signup",
	Tags:        []string{"Auth"},
	Summary:     "Start ZITADEL signup",
	Description: "Starts the browser-based ZITADEL registration flow and responds with `303 See Other` to the hosted signup UI. The hosted registration screen is requested with `prompt=create`.",
}

var zitadelCallbackOp = huma.Operation{
	OperationID: "auth-zitadel-callback",
	Method:      "GET",
	Path:        "/auth/zitadel/callback",
	Tags:        []string{"Auth"},
	Summary:     "Complete ZITADEL login",
	Description: "Completes the browser-based ZITADEL flow, links or provisions a local account, and responds with `303 See Other` back to the approved `return_to` URL. On success it appends `ticket=...`; on failure it appends `error` and `error_description`.",
}

var forceVerifyZitadelUserEmailOp = huma.Operation{
	OperationID: "admin-force-verify-zitadel-user-email",
	Method:      "POST",
	Path:        "/admin/zitadel/users/{userId}/email/force-verify",
	Tags:        []string{"Admin / ZITADEL"},
	Summary:     "Force-verify a ZITADEL user email",
	Description: "Administrative maintenance endpoint that uses a server-side ZITADEL admin token to verify an existing user's email. The backend first requests a verification code from ZITADEL and then immediately verifies the email with that code.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var exchangeOp = huma.Operation{
	OperationID: "auth-exchange",
	Method:      "POST",
	Path:        "/auth/exchange",
	Tags:        []string{"Auth"},
	Summary:     "Exchange a browser auth ticket",
	Description: "Consumes a short-lived one-time browser authentication ticket created by the ZITADEL browser callback and returns local access and refresh tokens.",
}

var refreshOp = huma.Operation{
	OperationID: "auth-refresh",
	Method:      "POST",
	Path:        "/auth/refresh",
	Tags:        []string{"Auth"},
	Summary:     "Refresh the current session",
	Description: "Exchanges a valid opaque refresh token for a fresh access token and refresh token pair. Refresh token rotation is one-time and detects reuse of older rotated tokens.",
}

var logoutOp = huma.Operation{
	OperationID: "auth-logout",
	Method:      "POST",
	Path:        "/auth/logout",
	Tags:        []string{"Auth"},
	Summary:     "Revoke the current session",
	Description: "Revokes the current refresh session so the client can no longer use the current or previously rotated refresh tokens from that session.",
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
