package app

import (
	accpg "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
	authapp "github.com/NikolayNam/collabsphere/internal/auth/application"
	authhttp "github.com/NikolayNam/collabsphere/internal/auth/delivery/http"
	authpg "github.com/NikolayNam/collabsphere/internal/auth/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/bcrypt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/tokens"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerAuthModule(api huma.API, db *gorm.DB) {
	accountRepo := accpg.NewAccountRepo(db)
	sessionRepo := authpg.NewSessionRepo(db)

	clk := clock.NewSystemClock()
	hasher := bcrypt.NewBcryptHasher()
	tokenGen := tokens.NewGenerator()

	// jwtManager := jwt.NewManager(...)
	_ = hasher
	_ = tokenGen

	authService := authapp.New(
		accountRepo,
		hasher,
		nil, // jwtManager
		tokenGen,
		sessionRepo,
		clk,
	)
	authHandler := authhttp.NewHandler(authService)
	authhttp.Register(api, authHandler)
}
