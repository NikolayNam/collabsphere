package system

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

var healthOp = huma.Operation{
	OperationID: "get-health",
	Method:      "GET",
	Path:        "/health",
	Tags:        []string{"System"},
	Summary:     "Check API health",
	Description: "Simple liveness and readiness probe for the API process.",
}

func Register(api huma.API) {
	huma.Register(api, healthOp, healthHandler)
}

func healthHandler(ctx context.Context, input *struct{}) (*HealthOutput, error) {
	resp := &HealthOutput{}
	resp.Body.Status = "ok"
	return resp, nil
}
