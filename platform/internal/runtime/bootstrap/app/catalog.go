package app

import (
	"fmt"

	catalogapp "github.com/NikolayNam/collabsphere/internal/catalog/application"
	catalogports "github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	cataloghttp "github.com/NikolayNam/collabsphere/internal/catalog/delivery/http"
	catalogpg "github.com/NikolayNam/collabsphere/internal/catalog/repository/postgres"
	memberspg "github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres"
	orgpg "github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	s3storage "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/storage/s3"
	uploadpg "github.com/NikolayNam/collabsphere/internal/uploads/repository/postgres"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerCatalogModule(api huma.API, db *gorm.DB, conf *config.Config) {
	repo := catalogpg.NewCatalogRepo(db)
	organizationRepo := orgpg.NewOrganizationRepo(db)
	membershipRepo := memberspg.NewMembershipRepo(db)
	txManager := dbtx.New(db)
	uploadRepo := uploadpg.NewRepo(db)
	clk := clock.NewSystemClock()

	var objectStorage catalogports.ObjectStorage
	if conf.Storage.S3.Enabled {
		client, err := s3storage.NewClient(conf.Storage.S3)
		if err != nil {
			panic(fmt.Errorf("init catalog s3 client: %w", err))
		}
		objectStorage = client
	}

	service := catalogapp.New(repo, organizationRepo, membershipRepo, txManager, clk, objectStorage, conf.Storage.S3.Bucket, uploadRepo)
	handler := cataloghttp.NewHandler(service)

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret for catalog: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	cataloghttp.Register(api, handler, jwtManager)
}
