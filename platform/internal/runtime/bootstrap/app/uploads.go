package app

import (
	"fmt"

	memberpg "github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	uploadapp "github.com/NikolayNam/collabsphere/internal/uploads/application"
	uploadhttp "github.com/NikolayNam/collabsphere/internal/uploads/delivery/http"
	uploadpg "github.com/NikolayNam/collabsphere/internal/uploads/repository/postgres"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerUploadsModule(api huma.API, db *gorm.DB, conf *config.Config) {
	repo := uploadpg.NewRepo(db)
	membershipRepo := memberpg.NewMembershipRepo(db)
	service := uploadapp.New(repo, membershipRepo)
	handler := uploadhttp.NewHandler(service)

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret for uploads: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	uploadhttp.Register(api, handler, jwtManager)
}
