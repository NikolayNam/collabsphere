package middleware

import (
	"context"
	"net/http"
	"strings"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
)

type principalCtxKey struct{}

var principalKey principalCtxKey

type AccessTokenVerifier interface {
	VerifyAccessToken(ctx context.Context, token string) (authdomain.Principal, error)
}

func WithPrincipal(ctx context.Context, p authdomain.Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

func PrincipalFromContext(ctx context.Context) authdomain.Principal {
	v := ctx.Value(principalKey)
	if v == nil {
		return authdomain.AnonymousPrincipal()
	}
	p, ok := v.(authdomain.Principal)
	if !ok {
		return authdomain.AnonymousPrincipal()
	}
	return p
}

func AuthOptional(verifier AccessTokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := strings.TrimSpace(r.Header.Get("Authorization"))
			if authz == "" {
				next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), authdomain.AnonymousPrincipal())))
				return
			}

			token := extractBearer(authz)
			if token == "" {
				next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), authdomain.AnonymousPrincipal())))
				return
			}

			p, err := verifier.VerifyAccessToken(r.Context(), token)
			if err != nil {
				next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), authdomain.AnonymousPrincipal())))
				return
			}

			next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), p)))
		})
	}
}

func AuthRequired(verifier AccessTokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		optional := AuthOptional(verifier)
		return optional(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := PrincipalFromContext(r.Context())
			if !p.Authenticated {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}

func extractBearer(v string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(v, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(v, prefix))
}
