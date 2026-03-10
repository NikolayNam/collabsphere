package app

import (
	"fmt"

	accpg "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
	authapp "github.com/NikolayNam/collabsphere/internal/auth/application"
	authports "github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	authhttp "github.com/NikolayNam/collabsphere/internal/auth/delivery/http"
	authpg "github.com/NikolayNam/collabsphere/internal/auth/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/bcrypt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/tokens"
	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	oidcprovider "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/security/oidc"
	zitadeladmin "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/security/zitadeladmin"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func registerAuthModule(router chi.Router, api huma.API, db *gorm.DB, conf *config.Config) {
	accountRepo := accpg.NewAccountRepo(db)
	sessionRepo := authpg.NewSessionRepo(db)
	externalIdentityRepo := authpg.NewExternalIdentityRepo(db)
	oidcStateRepo := authpg.NewOIDCStateRepo(db)
	oneTimeCodeRepo := authpg.NewOneTimeCodeRepo(db)
	txManager := dbtx.New(db)

	clk := clock.NewSystemClock()
	hasher := bcrypt.NewBcryptHasher()
	tokenGen := tokens.NewGenerator()

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	var oidcFlowProvider authports.OIDCProvider
	if conf.Auth.Zitadel.Enabled {
		oidcFlowProvider, err = oidcprovider.NewZitadelProvider(conf.Auth.Zitadel)
		if err != nil {
			panic(fmt.Errorf("build zitadel provider: %w", err))
		}
	}

	zitadelAdminClient, err := zitadeladmin.NewClient(conf.Auth.Zitadel)
	if err != nil {
		panic(fmt.Errorf("build zitadel admin client: %w", err))
	}

	authService := authapp.New(
		accountRepo,
		hasher,
		jwtManager,
		tokenGen,
		sessionRepo,
		clk,
		txManager,
		externalIdentityRepo,
		oidcStateRepo,
		oneTimeCodeRepo,
		oidcFlowProvider,
		zitadelAdminClient,
		conf.Auth.Zitadel.StateTTL,
		conf.Auth.Zitadel.NonceTTL,
		conf.Auth.BrowserTicketTTL,
	)
	authHandler := authhttp.NewHandler(authService, conf.Auth.PasswordLoginEnabled, zitadelAdminClient != nil, authhttp.BrowserFlowConfig{
		DefaultReturnURL:       conf.Auth.BrowserDefaultReturn,
		AllowedRedirectOrigins: conf.Auth.BrowserRedirectOriginList(),
	})
	authhttp.Register(router, api, authHandler, jwtManager)
}
