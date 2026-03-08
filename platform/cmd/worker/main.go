package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	accpg "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
	collabapp "github.com/NikolayNam/collabsphere/internal/collab/application"
	collabpg "github.com/NikolayNam/collabsphere/internal/collab/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/bootstrap"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/logger"
	jitsisec "github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jitsi"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/tokens"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/storage/s3"
	whisper "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/transcription/whisper"
)

func main() {
	conf := config.New()
	rootLog := logger.New(logger.Config{Level: slog.LevelInfo, Format: "json"})
	slog.SetDefault(rootLog)
	db := bootstrap.MustOpenGormDB(conf, rootLog.With("component", "db"))
	bootstrap.RegisterDBHooks(db)

	repo := collabpg.NewRepo(db)
	accountRepo := accpg.NewAccountRepo(db)
	clk := clock.NewSystemClock()
	tokenGen := tokens.NewGenerator()
	secret, err := conf.Auth.JWTSecretValue()
	if err != nil {
		log.Fatalf("load auth jwt secret: %v", err)
	}
	jwtManager := jwt.NewManager(secret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)
	var storageClient *s3.Client
	if conf.Storage.S3.Enabled {
		storageClient, err = s3.NewClient(conf.Storage.S3)
		if err != nil {
			log.Fatalf("init s3 client: %v", err)
		}
	}
	jitsiManager, err := jitsisec.NewManager(conf.Conference.Jitsi)
	if err != nil {
		log.Fatalf("init jitsi manager: %v", err)
	}
	transcriber, err := whisper.NewClient(conf.Transcription)
	if err != nil {
		log.Fatalf("init transcription client: %v", err)
	}
	service := collabapp.New(repo, accountRepo, storageClient, tokenGen, jwtManager, jitsiManager, clk, nil, transcriber, conf.APP.PublicBaseURL, conf.Storage.S3.Bucket, conf.Collab.GuestInviteTTL, conf.Auth.GuestAccessTTL)

	pollEvery := conf.Transcription.WorkerPollEvery
	if pollEvery <= 0 {
		pollEvery = 10 * time.Second
	}
	ticker := time.NewTicker(pollEvery)
	defer ticker.Stop()

	for {
		processed, err := service.ProcessNextTranscriptionJob(context.Background())
		if err != nil {
			slog.Error("transcription worker failed", "error", err.Error())
		}
		if processed {
			continue
		}
		<-ticker.C
	}
}
