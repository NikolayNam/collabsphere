package main

import (
	"log"
	"log/slog"

	"github.com/NikolayNam/collabsphere/internal/runtime/bootstrap"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/httpserver"
)

func main() {
	// 1) config (env + secrets + TZ)
	conf := config.New()

	// 2) build bootstrap (router + huma + module registration)
	application := bootstrap.New(conf)

	// 3) run http httpserver (timeouts + graceful shutdown)
	if err := httpserver.Run(application.Router, conf.APP.Address,
		httpserver.Options{
			ReadTimeout:       conf.APP.TimeoutRead,
			WriteTimeout:      conf.APP.TimeoutWrite,
			IdleTimeout:       conf.APP.TimeoutIdle,
			ReadHeaderTimeout: 5, // seconds
			ShutdownTimeout:   5, // seconds
			Logger:            slog.Default(),
		}); err != nil {
		log.Fatalf("httpserver failed: %v", err)
	}
}
