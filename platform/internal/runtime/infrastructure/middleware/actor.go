package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/actorctx"
)

const ActorHeader = "X-Actor-ID"

func Actor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimSpace(r.Header.Get(ActorHeader))
		if raw != "" {
			if id, err := uuid.Parse(raw); err == nil && id != uuid.Nil {
				r = r.WithContext(actorctx.WithActorID(r.Context(), id))
			}
		}
		next.ServeHTTP(w, r)
	})
}
