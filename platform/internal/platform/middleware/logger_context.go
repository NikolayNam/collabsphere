package middleware

import (
	"net/http"

	"log/slog"

	appLogger "github.com/NikolayNam/collabsphere-go/internal/platform/logger"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func LoggerContext(base *slog.Logger) func(http.Handler) http.Handler {
	if base == nil {
		panic("base logger is nil")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := chimw.GetReqID(r.Context())

			l := base.With(
				"request_id", reqID,
				"method", r.Method,
				"path", r.URL.Path,
			)

			ctx := appLogger.With(r.Context(), l)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
