package bootstrap

import (
	"github.com/go-chi/chi/v5"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
)

func NewAPI(router chi.Router, conf *config.Config) huma.API {
	humaerr.Install()

	cfg := huma.DefaultConfig(conf.APP.Title, conf.APP.Version)
	cfg.CreateHooks = nil
	if cfg.Components == nil {
		cfg.Components = &huma.Components{}
	}
	cfg.Components.Schemas = huma.NewMapRegistry("#/components/schemas/", newSchemaNamer())
	if cfg.Components.SecuritySchemes == nil {
		cfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{}
	}
	cfg.Components.SecuritySchemes["bearerAuth"] = &huma.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}

	// важно: чтобы Swagger/SDK знали, что API живёт под /v1
	cfg.Servers = []*huma.Server{
		{URL: "/api/v1"},
	}

	return humachi.New(router, cfg)
}
