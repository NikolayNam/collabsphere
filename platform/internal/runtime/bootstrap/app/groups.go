package app

import (
	"fmt"

	accpg "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
	groupsapp "github.com/NikolayNam/collabsphere/internal/groups/application"
	groupshttp "github.com/NikolayNam/collabsphere/internal/groups/delivery/http"
	groupspg "github.com/NikolayNam/collabsphere/internal/groups/repository/postgres"
	orgpg "github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerGroupsModule(api huma.API, db *gorm.DB, conf *config.Config) {
	groupRepo := groupspg.NewGroupRepo(db)
	accountRepo := accpg.NewAccountRepo(db)
	organizationRepo := orgpg.NewOrganizationRepo(db)
	txManager := dbtx.New(db)
	clk := clock.NewSystemClock()

	service := groupsapp.New(groupRepo, accountRepo, organizationRepo, txManager, clk)
	handler := groupshttp.NewHandler(service)

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret for groups: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	groupshttp.Register(api, handler, jwtManager)
}
