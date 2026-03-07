package app

import (
	"fmt"

	catalogapp "github.com/NikolayNam/collabsphere/internal/catalog/application"
	cataloghttp "github.com/NikolayNam/collabsphere/internal/catalog/delivery/http"
	catalogpg "github.com/NikolayNam/collabsphere/internal/catalog/repository/postgres"
	memberspg "github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres"
	orgpg "github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerCatalogModule(api huma.API, db *gorm.DB, conf *config.Config) {
	repo := catalogpg.NewCatalogRepo(db)
	organizationRepo := orgpg.NewOrganizationRepo(db)
	membershipRepo := memberspg.NewMembershipRepo(db)
	clk := clock.NewSystemClock()

	service := catalogapp.New(repo, organizationRepo, membershipRepo, clk)
	handler := cataloghttp.NewHandler(service)

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret for catalog: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	cataloghttp.Register(api, handler, jwtManager)
}
