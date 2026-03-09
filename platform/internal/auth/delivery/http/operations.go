package http

import "github.com/danielgtaylor/huma/v2"

var loginOp = huma.Operation{
	OperationID: "auth-login",
	Method:      "POST",
	Path:        "/auth/login",
	Tags:        []string{"Auth"},
	Summary:     "Login",
	Description: "Authenticates an account using email and password credentials and returns access and refresh tokens.",
}

var refreshOp = huma.Operation{
	OperationID: "auth-refresh",
	Method:      "POST",
	Path:        "/auth/refresh",
	Tags:        []string{"Auth"},
	Summary:     "Refresh tokens",
	Description: "Exchanges a valid refresh token for a fresh access token and refresh token pair.",
}

var logoutOp = huma.Operation{
	OperationID: "auth-logout",
	Method:      "POST",
	Path:        "/auth/logout",
	Tags:        []string{"Auth"},
	Summary:     "Logout",
	Description: "Revokes the current refresh session so the client can no longer use the associated refresh token.",
}

var meOp = huma.Operation{
	OperationID: "auth-me",
	Method:      "GET",
	Path:        "/auth/me",
	Tags:        []string{"Auth"},
	Summary:     "Get current user",
	Description: "Returns the authenticated principal and its current profile snapshot. For account principals this includes the current account profile fields.",
	Security: []map[string][]string{
		{"bearerAuth": {}},
	},
}
