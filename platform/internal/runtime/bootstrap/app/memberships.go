package app

import (
	membershipsApp "github.com/NikolayNam/collabsphere/internal/memberships/application"
	membershipsHTTP "github.com/NikolayNam/collabsphere/internal/memberships/delivery/http"
	membershipsPG "github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres"

	orgPG "github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerMembershipsModule(api huma.API, db *gorm.DB) {
	membershipRepo := membershipsPG.NewMembershipRepo(db)

	organizationRepo := orgPG.NewOrganizationRepo(db)

	clk := clock.NewSystemClock()

	membershipService := membershipsApp.New(membershipRepo, organizationRepo, clk)
	membershipHandler := membershipsHTTP.NewHandler(membershipService)

	membershipsHTTP.Register(api, membershipHandler)
}
