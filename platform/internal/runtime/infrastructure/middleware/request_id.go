package middleware

import (
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/requestctx"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := chimw.GetReqID(r.Context())
		if requestID == "" {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r.WithContext(requestctx.WithRequestID(r.Context(), requestID)))
	})
}
