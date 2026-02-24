package bootstrap

import (
	"log/slog"
	"time"

	"github.com/NikolayNam/collabsphere-go/internal/platform/middleware"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func newRouter(log *slog.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(chimw.RequestSize(1 << 20))

	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.LoggerContext(log))
	r.Use(middleware.AccessLog())

	r.Use(middleware.RateLimit)

	return r
}
