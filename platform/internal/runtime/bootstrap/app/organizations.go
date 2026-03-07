package app

import (
	"fmt"

	catalogpg "github.com/NikolayNam/collabsphere/internal/catalog/repository/postgres"
	memberspg "github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/organizations/application"
	orghttp "github.com/NikolayNam/collabsphere/internal/organizations/delivery/http"
	orgpg "github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerOrganzationsModule(api huma.API, db *gorm.DB, conf *config.Config) {
	organizationRepo := orgpg.NewOrganizationRepo(db)
	membershipRepo := memberspg.NewMembershipRepo(db)
	categoryRepo := catalogpg.NewProductCategoryRepo(db)
	txManager := dbtx.New(db)
	clk := clock.NewSystemClock()

	organizationService := application.New(organizationRepo, membershipRepo, categoryRepo, txManager, clk)
	organizationHandler := orghttp.NewHandler(organizationService)

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret for organizations: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	orghttp.Register(api, organizationHandler, jwtManager)
}
