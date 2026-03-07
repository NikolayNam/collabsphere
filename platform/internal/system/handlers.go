package system

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API) {
	huma.Get(api, "/health", healthHandler)
}

func healthHandler(ctx context.Context, input *struct{}) (*HealthOutput, error) {
	resp := &HealthOutput{}
	resp.Body.Status = "ok"
	return resp, nil
}
