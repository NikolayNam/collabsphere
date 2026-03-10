package http

import "github.com/danielgtaylor/huma/v2"

var getUploadOp = huma.Operation{
	OperationID: "get-upload",
	Method:      "GET",
	Path:        "/uploads/{id}",
	Tags:        []string{"Uploads"},
	Summary:     "Get upload session state",
	Description: "Returns the current state of an upload session, including object metadata, failure details, and any linked result entity.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
