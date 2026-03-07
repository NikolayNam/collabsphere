package http

import "github.com/danielgtaylor/huma/v2"

var loginOp = huma.Operation{
	OperationID: "auth-login",
	Method:      "POST",
	Path:        "/auth/login",
	Tags:        []string{"Auth"},
	Summary:     "Login",
}

var refreshOp = huma.Operation{
	OperationID: "auth-refresh",
	Method:      "POST",
	Path:        "/auth/refresh",
	Tags:        []string{"Auth"},
	Summary:     "Refresh tokens",
}

var logoutOp = huma.Operation{
	OperationID: "auth-logout",
	Method:      "POST",
	Path:        "/auth/logout",
	Tags:        []string{"Auth"},
	Summary:     "Logout",
}

var meOp = huma.Operation{
	OperationID: "auth-me",
	Method:      "GET",
	Path:        "/auth/me",
	Tags:        []string{"Auth"},
	Summary:     "Get current user",
}
