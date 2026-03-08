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
	memberspg "github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres"
	orgapp "github.com/NikolayNam/collabsphere/internal/organizations/application"
	orgpg "github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/bootstrap"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/logger"
	jitsisec "github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jitsi"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/tokens"
	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	generichttp "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/documentanalysis/generichttp"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/storage/s3"
	whisper "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/transcription/whisper"
)

func main() {
	conf := config.New()
	rootLog := logger.New(logger.Config{Level: slog.LevelInfo, Format: "json"})
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
	jitsiManager, err := jitsisec.NewManager(conf.Conference.Jitsi)
	if err != nil {
		log.Fatalf("init jitsi manager: %v", err)
	}
	transcriber, err := whisper.NewClient(conf.Transcription)
	if err != nil {
		log.Fatalf("init transcription client: %v", err)
	}
	collabService := collabapp.New(collabRepo, accountRepo, storageClient, tokenGen, jwtManager, jitsiManager, clk, nil, transcriber, conf.APP.PublicBaseURL, conf.Storage.S3.Bucket, conf.Collab.GuestInviteTTL, conf.Auth.GuestAccessTTL)

	organizationRepo := orgpg.NewOrganizationRepo(db)
	membershipRepo := memberspg.NewMembershipRepo(db)
	categoryRepo := catalogpg.NewProductCategoryRepo(db)
	txManager := dbtx.New(db)
	documentAnalyzer, err := generichttp.NewClient(conf.DocumentAnalysis)
	if err != nil {
		log.Fatalf("init document analysis client: %v", err)
	}
	organizationService := orgapp.New(organizationRepo, membershipRepo, categoryRepo, txManager, clk, storageClient, conf.Storage.S3.Bucket, documentAnalyzer, conf.DocumentAnalysis.Provider)

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
