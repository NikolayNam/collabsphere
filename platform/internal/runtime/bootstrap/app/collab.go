package app

import (
	"fmt"

	accpg "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
	collabapp "github.com/NikolayNam/collabsphere/internal/collab/application"
	collabhttp "github.com/NikolayNam/collabsphere/internal/collab/delivery/http"
	"github.com/NikolayNam/collabsphere/internal/collab/realtime"
	collabpg "github.com/NikolayNam/collabsphere/internal/collab/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	jitsisec "github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jitsi"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/tokens"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/storage/s3"
	whisper "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/transcription/whisper"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func registerCollabModule(api huma.API, router chi.Router, db *gorm.DB, conf *config.Config) {
	repo := collabpg.NewRepo(db)
	accountRepo := accpg.NewAccountRepo(db)
	clk := clock.NewSystemClock()
	tokenGen := tokens.NewGenerator()
	broker := realtime.NewBroker()

	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		panic(fmt.Errorf("load auth jwt secret for collab: %w", err))
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	var storageClient *s3.Client
	if conf.Storage.S3.Enabled {
		storageClient, err = s3.NewClient(conf.Storage.S3)
		if err != nil {
			panic(fmt.Errorf("init collab storage client: %w", err))
		}
	}

	conferenceProvider := conf.Conference.ProviderValue()
	switch conferenceProvider {
	case "jitsi", "mediasoup":
	default:
		panic(fmt.Errorf("unsupported conference provider: %s", conferenceProvider))
	}

	var jitsiManager *jitsisec.Manager
	if conf.Conference.Jitsi.Enabled {
		jitsiManager, err = jitsisec.NewManager(conf.Conference.Jitsi)
		if err != nil {
			panic(fmt.Errorf("init jitsi manager: %w", err))
		}
	}
	if conferenceProvider == "jitsi" && jitsiManager == nil {
		panic(fmt.Errorf("conference provider jitsi requires JITSI_ENABLED=true"))
	}

	transcriber, err := whisper.NewClient(conf.Transcription)
	if err != nil {
		panic(fmt.Errorf("init transcription client: %w", err))
	}

	service := collabapp.New(repo, accountRepo, storageClient, tokenGen, jwtManager, jitsiManager, clk, broker, transcriber, conferenceProvider, conf.APP.PublicBaseURL, conf.Storage.S3.Bucket, conf.Collab.GuestInviteTTL, conf.Auth.GuestAccessTTL)
	handler := collabhttp.NewHandler(service)
	collabhttp.Register(api, handler, jwtManager)
	if router != nil {
		router.Get("/ws/collab", collabhttp.NewWebSocketHandler(service, jwtManager, broker).ServeHTTP)
	}
}
