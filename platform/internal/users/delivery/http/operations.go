package http

import "github.com/danielgtaylor/huma/v2"

var createUserOp = huma.Operation{
	OperationID: "create-user",
	Method:      "POST",
	Path:        "/users",
	Tags:        []string{"Users"},
	Summary:     "Create user",
}
