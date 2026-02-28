package bootstrap

import (
	"log/slog"

	"gorm.io/gorm"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/logger"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/bcrypt"

	"github.com/NikolayNam/collabsphere/internal/accounts/application"
	"github.com/NikolayNam/collabsphere/internal/accounts/delivery/http"
	"github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
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

	router := newRouter(httpLog)
	api := newAPI(router, conf)

	db := mustOpenGormDB(conf, dbLog)
	registerDBHooks(db)

	registerPlatform(api)
	registerAccountsModule(api, db /*, rootLog*/)

	appLog.Info("application bootstrapped",
		"event", "app.bootstrap.completed",
	)
	return &App{
		Router: router,
		API:    api,
	}
}

func registerPlatform(api huma.API) {
	system.Register(api)
}

func registerAccountsModule(api huma.API, db *gorm.DB) {
	accountRepo := postgres.NewAccountRepo(db)

	hasher := bcrypt.NewBcryptHasher()
	clk := clock.NewSystemClock()

	accountService := application.New(accountRepo, hasher, clk)
	accountHandler := http.NewHandler(accountService)

	http.Register(api, accountHandler)
}
