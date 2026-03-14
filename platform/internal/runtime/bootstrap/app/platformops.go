package app

import (
	"fmt"

	accpg "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
	authports "github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	platformapp "github.com/NikolayNam/collabsphere/internal/platformops/application"
	platformhttp "github.com/NikolayNam/collabsphere/internal/platformops/delivery/http"
	platformpg "github.com/NikolayNam/collabsphere/internal/platformops/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	zitadeladmin "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/security/zitadeladmin"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerPlatformOpsModule(api huma.API, db *gorm.DB, conf *config.Config) {
	accountRepo := accpg.NewAccountRepo(db)
	autoGrantCfg, err := conf.Auth.PlatformAutoGrantRules()
	if err != nil {
		panic(fmt.Errorf("load platform auto-grant rules: %w", err))
	}
	repo := platformpg.NewRepo(db, platformpg.WithBootstrapAutoGrantRules(buildBootstrapAutoGrantRules(autoGrantCfg)))
	txManager := dbtx.New(db)
	clk := clock.NewSystemClock()

	bootstrapAccountIDs, err := conf.Auth.PlatformBootstrapAccountUUIDs()
	if err != nil {
		panic(fmt.Errorf("parse AUTH_PLATFORM_BOOTSTRAP_ACCOUNT_IDS: %w", err))
	}

	zitadelAdminClient, err := zitadeladmin.NewClient(conf.Auth.Zitadel)
	if err != nil {
		panic(fmt.Errorf("build platform zitadel admin client: %w", err))
	}
	var zitadelAdmin authports.ZitadelAdminClient
	if zitadelAdminClient != nil {
		zitadelAdmin = zitadelAdminClient
	}

	service := platformapp.New(repo, repo, repo, accountRepo, repo, repo, repo, clk, txManager, zitadelAdmin, bootstrapAccountIDs)
	handler := platformhttp.NewHandler(service)

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret for platform control-plane: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	platformhttp.Register(api, handler, jwtManager)
}
