package app

import (
	"fmt"

	marketapp "github.com/NikolayNam/collabsphere/internal/marketplace/application"
	markethttp "github.com/NikolayNam/collabsphere/internal/marketplace/delivery/http"
	marketpg "github.com/NikolayNam/collabsphere/internal/marketplace/repository/postgres"
	memberspg "github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerMarketplaceModule(api huma.API, db *gorm.DB, conf *config.Config) {
	repo := marketpg.NewRepo(db)
	memberships := memberspg.NewMembershipRepo(db)
	txm := dbtx.New(db)
	clk := clock.NewSystemClock()

	service := marketapp.New(repo, memberships, txm, clk)
	handler := markethttp.NewHandler(service)

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret for marketplace: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)
	markethttp.Register(api, handler, jwtManager)
}
