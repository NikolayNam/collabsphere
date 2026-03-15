package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	accpg "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
	catalogpg "github.com/NikolayNam/collabsphere/internal/catalog/repository/postgres"
	collabapp "github.com/NikolayNam/collabsphere/internal/collab/application"
	collabpg "github.com/NikolayNam/collabsphere/internal/collab/repository/postgres"
	membershipsApp "github.com/NikolayNam/collabsphere/internal/memberships/application"
	memberspg "github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres"
	orgapp "github.com/NikolayNam/collabsphere/internal/organizations/application"
	orgpg "github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/bootstrap"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/logger"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/tokens"
	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	generichttp "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/documentanalysis/generichttp"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/storage/s3"
	whisper "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/transcription/whisper"
	uploadpg "github.com/NikolayNam/collabsphere/internal/uploads/repository/postgres"
)

func main() {
	conf := config.NewFor(config.ProfileWorker)
	rootLog := logger.New(logger.Config{
		Level:  slog.LevelInfo,
		Format: "json",
		Fields: []any{
			"service", "worker",
			"env", conf.APP.NormalizedEnvironment(),
		},
	})
	slog.SetDefault(rootLog)
	db := bootstrap.MustOpenGormDB(conf, rootLog.With("component", "db"))
	bootstrap.RegisterDBHooks(db)

	clk := clock.NewSystemClock()
	var storageClient *s3.Client
	var err error
	if conf.Storage.S3.Enabled {
		storageClient, err = s3.NewClient(conf.Storage.S3)
		if err != nil {
			log.Fatalf("init s3 client: %v", err)
		}
	}

	collabRepo := collabpg.NewRepo(db)
	accountRepo := accpg.NewAccountRepo(db)
	tokenGen := tokens.NewGenerator()
	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		log.Fatalf("load auth jwt secret: %v", err)
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)
	conferenceProvider := conf.Conference.ProviderValue()
	if conferenceProvider != "mediasoup" {
		log.Fatalf("unsupported conference provider: %s", conferenceProvider)
	}
	transcriber, err := whisper.NewClient(conf.Transcription)
	if err != nil {
		log.Fatalf("init transcription client: %v", err)
	}
	collabService := collabapp.New(collabRepo, accountRepo, storageClient, tokenGen, jwtManager, clk, nil, transcriber, conferenceProvider, conf.APP.PublicBaseURL, conf.Storage.S3.Bucket, conf.Collab.GuestInviteTTL, conf.Auth.GuestAccessTTL)

	organizationRepo := orgpg.NewOrganizationRepo(db)
	membershipRepo := memberspg.NewMembershipRepo(db)
	roleRepo := memberspg.NewOrganizationRoleRepo(db)
	roleResolver := membershipsApp.NewRoleResolverAdapter(roleRepo)
	categoryRepo := catalogpg.NewProductCategoryRepo(db)
	catalogRepo := catalogpg.NewCatalogRepo(db)
	txManager := dbtx.New(db)
	uploadRepo := uploadpg.NewRepo(db)
	documentAnalyzer, err := generichttp.NewClient(conf.DocumentAnalysis)
	if err != nil {
		log.Fatalf("init document analysis client: %v", err)
	}
	organizationService := orgapp.New(organizationRepo, membershipRepo, roleResolver, categoryRepo, catalogRepo, txManager, clk, storageClient, conf.Storage.S3.Bucket, documentAnalyzer, conf.DocumentAnalysis.Provider, uploadRepo)

	pollEvery := smallestPositiveDuration(conf.Transcription.WorkerPollEvery, conf.DocumentAnalysis.WorkerPollEvery)
	if pollEvery <= 0 {
		pollEvery = 10 * time.Second
	}
	ticker := time.NewTicker(pollEvery)
	defer ticker.Stop()

	for {
		processed, err := collabService.ProcessNextTranscriptionJob(context.Background())
		if err != nil {
			slog.Error("transcription worker failed", "error", err.Error())
		}
		if processed {
			continue
		}

		processed, err = organizationService.ProcessNextLegalDocumentAnalysisJob(context.Background())
		if err != nil {
			slog.Error("legal document analysis worker failed", "error", err.Error())
		}
		if processed {
			continue
		}

		<-ticker.C
	}
}

func smallestPositiveDuration(values ...time.Duration) time.Duration {
	var out time.Duration
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if out == 0 || value < out {
			out = value
		}
	}
	return out
}
