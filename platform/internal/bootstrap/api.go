package bootstrap

import (
	"github.com/NikolayNam/collabsphere-go/internal/platform/config"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

func newAPI(router chi.Router, conf *config.Config) huma.API {
	cfg := huma.DefaultConfig(conf.APP.Title, conf.APP.Version)
	cfg.CreateHooks = nil

	return humachi.New(router, cfg)
}
