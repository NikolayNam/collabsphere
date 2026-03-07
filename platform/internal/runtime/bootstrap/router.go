package bootstrap

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
)

func NewRouter(httpLog *slog.Logger) chi.Router {
	humaerr.Install()
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)

	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.RateLimit)
	r.Use(middleware.Actor)
	r.Use(middleware.AccessLog(httpLog))

	return r
}
