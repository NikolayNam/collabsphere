package bootstrap

import (
	"github.com/NikolayNam/collabsphere-go/internal/platform/config"
	appLogger "github.com/NikolayNam/collabsphere-go/internal/platform/logger"
	"gorm.io/gorm"

	systemapp "github.com/NikolayNam/collabsphere-go/internal/system"
	usersapp "github.com/NikolayNam/collabsphere-go/internal/users/application"
	usershttp "github.com/NikolayNam/collabsphere-go/internal/users/delivery/http"
	usersrepo "github.com/NikolayNam/collabsphere-go/internal/users/repository/postgres"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
)

type App struct {
	Router chi.Router
	API    huma.API
}

func New(conf *config.Config) *App {
	log := appLogger.New()

	router := newRouter(log)
	api := newAPI(router, conf)

	db := mustOpenGormDB(conf, log)
	registerDBHooks(db)

	registerPlatform(api)
	registerUsersModule(api, db)

	return &App{
		Router: router,
		API:    api,
	}
}

func registerPlatform(api huma.API) {
	systemapp.Register(api)
}

func registerUsersModule(api huma.API, db *gorm.DB) {
	userRepo := usersrepo.NewUserRepo(db)
	userApp := usersapp.New(userRepo)
	userHandler := usershttp.NewHandler(userApp)

	usershttp.Register(api, userHandler)
}
