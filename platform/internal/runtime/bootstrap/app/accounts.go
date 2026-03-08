package app

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/bcrypt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	s3storage "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/storage/s3"
	"github.com/danielgtaylor/huma/v2"

	"github.com/NikolayNam/collabsphere/internal/accounts/application"
	accports "github.com/NikolayNam/collabsphere/internal/accounts/application/ports"
	acchttp "github.com/NikolayNam/collabsphere/internal/accounts/delivery/http"
	"github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
)

func registerAccountsModule(api huma.API, db *gorm.DB, conf *config.Config) {
	accountRepo := postgres.NewAccountRepo(db)

	hasher := bcrypt.NewBcryptHasher()
	clk := clock.NewSystemClock()

	var objectStorage accports.ObjectStorage
	if conf.Storage.S3.Enabled {
		client, err := s3storage.NewClient(conf.Storage.S3)
		if err != nil {
			panic(fmt.Errorf("init accounts s3 client: %w", err))
		}
		objectStorage = client
	}

	accountService := application.New(accountRepo, hasher, clk, objectStorage, conf.Storage.S3.Bucket)
	accountHandler := acchttp.NewHandler(accountService)

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret for accounts: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	acchttp.Register(api, accountHandler, jwtManager)
}
