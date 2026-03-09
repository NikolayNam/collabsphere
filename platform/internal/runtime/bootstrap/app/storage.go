package app

import (
	"fmt"

	collabpg "github.com/NikolayNam/collabsphere/internal/collab/repository/postgres"
	memberpg "github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	s3storage "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/storage/s3"
	storageapp "github.com/NikolayNam/collabsphere/internal/storage/application"
	storagehttp "github.com/NikolayNam/collabsphere/internal/storage/delivery/http"
	storagepg "github.com/NikolayNam/collabsphere/internal/storage/repository/postgres"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func registerStorageModule(api huma.API, db *gorm.DB, conf *config.Config) {
	repo := storagepg.NewRepo(db)
	membershipRepo := memberpg.NewMembershipRepo(db)
	collabRepo := collabpg.NewRepo(db)

	var objectStorage storageapp.ObjectStorage
	if conf.Storage.S3.Enabled {
		client, err := s3storage.NewClient(conf.Storage.S3)
		if err != nil {
			panic(fmt.Errorf("init storage s3 client: %w", err))
		}
		objectStorage = client
	}

	service := storageapp.New(repo, membershipRepo, collabRepo, objectStorage)
	handler := storagehttp.NewHandler(service)

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret for storage: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	storagehttp.Register(api, handler, jwtManager)
}
