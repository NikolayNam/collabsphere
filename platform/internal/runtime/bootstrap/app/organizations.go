package app

// imports:
import (
	"github.com/NikolayNam/collabsphere/internal/organizations/application"
	"github.com/NikolayNam/collabsphere/internal/organizations/delivery/http"
	"github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerOrganzationsModule(api huma.API, db *gorm.DB) {
	organizationRepo := postgres.NewOrganizationRepo(db)
	clk := clock.NewSystemClock()

	organizationService := application.New(organizationRepo, clk)
	organizationHandler := http.NewHandler(organizationService)
	http.Register(api, organizationHandler)
}
