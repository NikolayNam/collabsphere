package app

import (
	"gorm.io/gorm"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/bcrypt"
	"github.com/danielgtaylor/huma/v2"

	"github.com/NikolayNam/collabsphere/internal/accounts/application"
	"github.com/NikolayNam/collabsphere/internal/accounts/delivery/http"
	"github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
)

func registerAccountsModule(api huma.API, db *gorm.DB) {
	accountRepo := postgres.NewAccountRepo(db)

	hasher := bcrypt.NewBcryptHasher()
	clk := clock.NewSystemClock()

	accountService := application.New(accountRepo, hasher, clk)
	accountHandler := http.NewHandler(accountService)

	http.Register(api, accountHandler)
}
