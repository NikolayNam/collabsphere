package app

import (
	"fmt"

	membershipsApp "github.com/NikolayNam/collabsphere/internal/memberships/application"
	membershipsHTTP "github.com/NikolayNam/collabsphere/internal/memberships/delivery/http"
	membershipsPG "github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres"
	orgPG "github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerMembershipsModule(api huma.API, db *gorm.DB, conf *config.Config) {
	membershipRepo := membershipsPG.NewMembershipRepo(db)
	organizationRepo := orgPG.NewOrganizationRepo(db)
	clk := clock.NewSystemClock()

	membershipService := membershipsApp.New(membershipRepo, organizationRepo, clk)
	membershipHandler := membershipsHTTP.NewHandler(membershipService)

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret for memberships: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	membershipsHTTP.Register(api, membershipHandler, jwtManager)
}
