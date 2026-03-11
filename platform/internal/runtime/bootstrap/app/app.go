package app

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"

	"github.com/NikolayNam/collabsphere/internal/runtime/bootstrap"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/logger"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/authcallback"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/docsportal"
	"github.com/NikolayNam/collabsphere/internal/system"
)

type App struct {
	Router chi.Router
	API    huma.API
}

func New(conf *config.Config) *App {
	rootLog := logger.New(logger.Config{
		Level:     slog.LevelInfo,
		AddSource: false,
		Format:    "json",
	})

	slog.SetDefault(rootLog)

	appLog := rootLog.With("component", "app")
	httpLog := rootLog.With("component", "http")
	dbLog := rootLog.With("component", "db")

	router := bootstrap.NewRouter(httpLog)

	apiV1 := chi.NewRouter()
	router.Mount("/v1", apiV1)

	api := bootstrap.NewAPI(apiV1, conf)
	bootstrap.RegisterScalarDocs(apiV1, conf.APP.Title, "/v1/openapi.json")
	authcallback.Register(router, conf.APP.Title)
	docsportal.Register(router, conf.APP.Title)
	router.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/v1/openapi.yaml", http.StatusTemporaryRedirect)
	})
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/v1/health", http.StatusTemporaryRedirect)
	})

	db := bootstrap.MustOpenGormDB(conf, dbLog)
	bootstrap.RegisterDBHooks(db)

	registerPlatform(api)
	registerPlatformOpsModule(api, db, conf)
	registerAccountsModule(api, db, conf)
	registerOrganzationsModule(api, db, conf)
	registerMembershipsModule(api, db, conf)
	registerCatalogModule(api, db, conf)
	registerGroupsModule(api, db, conf)
	registerCollabModule(api, router, db, conf)
	registerStorageModule(api, db, conf)
	registerAuthModule(apiV1, api, db, conf)

	appLog.Info("application bootstrapped",
		"event", "app.bootstrap.completed",
	)
	return &App{Router: router, API: api}
}

func registerPlatform(api huma.API) {
	system.Register(api)
}
