package app

import (
	"fmt"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	tenantsapp "github.com/NikolayNam/collabsphere/internal/tenants/application"
	tenantshttp "github.com/NikolayNam/collabsphere/internal/tenants/delivery/http"
	tenantspg "github.com/NikolayNam/collabsphere/internal/tenants/repository/postgres"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerTenantsModule(api huma.API, db *gorm.DB, conf *config.Config) {
	repo := tenantspg.NewTenantRepo(db)
	txm := dbtx.New(db)
	clk := clock.NewSystemClock()

	service := tenantsapp.New(repo, txm, clk)
	handler := tenantshttp.NewHandler(service)

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret for tenants: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)
	tenantshttp.Register(api, handler, jwtManager)
}
