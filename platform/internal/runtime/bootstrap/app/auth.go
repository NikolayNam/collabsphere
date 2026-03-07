package app

import (
	"fmt"

	accpg "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
	authapp "github.com/NikolayNam/collabsphere/internal/auth/application"
	authhttp "github.com/NikolayNam/collabsphere/internal/auth/delivery/http"
	authpg "github.com/NikolayNam/collabsphere/internal/auth/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/bcrypt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/tokens"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerAuthModule(api huma.API, db *gorm.DB, conf *config.Config) {
	accountRepo := accpg.NewAccountRepo(db)
	sessionRepo := authpg.NewSessionRepo(db)

	clk := clock.NewSystemClock()
	hasher := bcrypt.NewBcryptHasher()
	tokenGen := tokens.NewGenerator()

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	authService := authapp.New(
		accountRepo,
		hasher,
		jwtManager,
		tokenGen,
		sessionRepo,
		clk,
	)
	authHandler := authhttp.NewHandler(authService)
	authhttp.Register(api, authHandler, jwtManager)
}
