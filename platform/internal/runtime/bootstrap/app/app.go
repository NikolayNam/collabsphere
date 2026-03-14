package app

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"

	"github.com/NikolayNam/collabsphere/internal/runtime/bootstrap"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/logger"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/authcallback"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/docsportal"
	runtimemetrics "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/metrics"
	"github.com/NikolayNam/collabsphere/internal/system"
	"gorm.io/gorm"
)

type App struct {
	Router chi.Router
	API    huma.API
}

func New(conf *config.Config) *App {
	return buildApp(conf, bootstrap.MustOpenGormDB, true)
}

func NewContracts(conf *config.Config) *App {
	return buildApp(conf, func(conf *config.Config, dbLog *slog.Logger) *gorm.DB {
		return bootstrap.MustOpenNoopGormDB(dbLog)
	}, false)
}

func buildApp(conf *config.Config, openDB func(*config.Config, *slog.Logger) *gorm.DB, registerDBHooks bool) *App {
	rootLog := logger.New(logger.Config{
		Level:     logger.ParseLevel(conf.APP.LogLevel),
		AddSource: false,
		Format:    "json",
		Fields: []any{
			"service", "api",
			"env", conf.APP.NormalizedEnvironment(),
		},
	})

	slog.SetDefault(rootLog)

	appLog := rootLog.With("component", "app")
	httpLog := rootLog.With("component", "http")
	dbLog := rootLog.With("component", "db")

	quietPaths := []string{"/health", "/ready", "/v1/health", "/v1/ready"}
	routerOptions := bootstrap.RouterOptions{
		AccessLogQuietPaths: quietPaths,
	}
	var httpMetrics *runtimemetrics.HTTP
	if conf.APP.MetricsEnabled {
		metricsPath := conf.APP.MetricsRoutePath()
		httpMetrics = runtimemetrics.NewHTTP(runtimemetrics.HTTPOptions{
			SkippedPaths: []string{"/health", "/ready", "/v1/health", "/v1/ready", metricsPath},
		})
		routerOptions.AccessLogQuietPaths = append(routerOptions.AccessLogQuietPaths, metricsPath)
		routerOptions.HTTPMetrics = httpMetrics.Middleware()
		appLog.Info("http metrics enabled",
			"event", "app.metrics.enabled",
			"path", metricsPath,
		)
	}

	router := bootstrap.NewRouter(httpLog, routerOptions)

	apiV1 := chi.NewRouter()
	router.Mount("/v1", apiV1)
	if httpMetrics != nil {
		metricsPath := conf.APP.MetricsRoutePath()
		if strings.HasPrefix(metricsPath, "/v1/") {
			apiV1.Handle(strings.TrimPrefix(metricsPath, "/v1"), httpMetrics.Handler())
		} else {
			router.Handle(metricsPath, httpMetrics.Handler())
		}
	}

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
	router.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/v1/ready", http.StatusTemporaryRedirect)
	})

	db := openDB(conf, dbLog)
	if registerDBHooks {
		bootstrap.RegisterDBHooks(db)
	}

	registerPlatform(api, system.ReadyFunc(func(ctx context.Context) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	}))
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

func registerPlatform(api huma.API, checker system.ReadyChecker) {
	system.Register(api, checker)
}
