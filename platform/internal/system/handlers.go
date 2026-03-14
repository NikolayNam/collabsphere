package system

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/danielgtaylor/huma/v2"
)

var healthOp = huma.Operation{
	OperationID: "get-health",
	Method:      "GET",
	Path:        "/health",
	Tags:        []string{"System"},
	Summary:     "Check API health",
	Description: "Simple liveness probe for the API process.",
}

var readyOp = huma.Operation{
	OperationID: "get-ready",
	Method:      "GET",
	Path:        "/ready",
	Tags:        []string{"System"},
	Summary:     "Check API readiness",
	Description: "Readiness probe that verifies critical runtime dependencies such as the primary database connection.",
}

type ReadyChecker interface {
	Ready(context.Context) error
}

type ReadyFunc func(context.Context) error

func (f ReadyFunc) Ready(ctx context.Context) error {
	return f(ctx)
}

func Register(api huma.API, checker ReadyChecker) {
	huma.Register(api, healthOp, healthHandler)
	huma.Register(api, readyOp, readyHandler(checker))
}

func healthHandler(ctx context.Context, input *struct{}) (*HealthOutput, error) {
	resp := &HealthOutput{}
	resp.Body.Status = "ok"
	return resp, nil
}

func readyHandler(checker ReadyChecker) func(context.Context, *struct{}) (*ReadyOutput, error) {
	return func(ctx context.Context, input *struct{}) (*ReadyOutput, error) {
		if checker != nil {
			if err := checker.Ready(ctx); err != nil {
				return nil, fault.Unavailable("Readiness check failed", fault.WithCause(err), fault.Code("SYSTEM_NOT_READY"))
			}
		}
		resp := &ReadyOutput{}
		resp.Body.Status = "ready"
		return resp, nil
	}
}
