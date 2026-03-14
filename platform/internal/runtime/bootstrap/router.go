package bootstrap

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
)

type RouterOptions struct {
	AccessLogQuietPaths []string
	HTTPMetrics         func(http.Handler) http.Handler
	RateLimit           func(http.Handler) http.Handler
}

func NewRouter(httpLog *slog.Logger, options RouterOptions) chi.Router {
	humaerr.Install()
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(middleware.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)

	r.Use(middleware.SecurityHeaders)
	rateLimitMiddleware := options.RateLimit
	if rateLimitMiddleware == nil {
		rateLimitMiddleware = middleware.NewRateLimit(middleware.RateLimitOptions{})
	}
	r.Use(rateLimitMiddleware)
	r.Use(middleware.Actor)
	if options.HTTPMetrics != nil {
		r.Use(options.HTTPMetrics)
	}
	r.Use(middleware.AccessLog(httpLog, middleware.AccessLogOptions{QuietPaths: options.AccessLogQuietPaths}))

	return r
}
