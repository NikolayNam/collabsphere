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
