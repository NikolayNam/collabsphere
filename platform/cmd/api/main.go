package main

import (
	"log"

	configs "github.com/NikolayNam/collabsphere-go/internal/platform/config"

	boot "github.com/NikolayNam/collabsphere-go/internal/bootstrap"
	srv "github.com/NikolayNam/collabsphere-go/internal/platform/httpserver"
)

func main() {
	// 1) config (env + secrets + TZ)
	conf := configs.New()

	// 2) build bootstrap (router + huma + module registration)
	application := boot.New(conf)

	// 3) run http httpserver (timeouts + graceful shutdown)
	if err := srv.Run(application.Router, conf.APP.Address,
		srv.Options{
			ReadTimeout:       conf.APP.TimeoutRead,
			WriteTimeout:      conf.APP.TimeoutWrite,
			IdleTimeout:       conf.APP.TimeoutIdle,
			ReadHeaderTimeout: 5, // seconds
			ShutdownTimeout:   5, // seconds
		}); err != nil {
		log.Fatalf("httpserver failed: %v", err)
	}
}
