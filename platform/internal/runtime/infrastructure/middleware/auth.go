package middleware

import (
	"context"
	"net/http"
	"strings"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/actorctx"
	"github.com/danielgtaylor/huma/v2"
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
			principal := authenticatePrincipal(r.Context(), strings.TrimSpace(r.Header.Get("Authorization")), verifier)
			next.ServeHTTP(w, r.WithContext(withActor(WithPrincipal(r.Context(), principal), principal)))
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

func HumaAuthOptional(verifier AccessTokenVerifier) func(huma.Context, func(huma.Context)) {
	return func(hctx huma.Context, next func(huma.Context)) {
		principal := authenticatePrincipal(hctx.Context(), strings.TrimSpace(hctx.Header("Authorization")), verifier)
		ctx := withActor(WithPrincipal(hctx.Context(), principal), principal)
		next(huma.WithContext(hctx, ctx))
	}
}

func authenticatePrincipal(ctx context.Context, authz string, verifier AccessTokenVerifier) authdomain.Principal {
	principal := authdomain.AnonymousPrincipal()
	if verifier == nil || authz == "" {
		return principal
	}

	token := extractBearer(authz)
	if token == "" {
		return principal
	}

	verified, err := verifier.VerifyAccessToken(ctx, token)
	if err != nil {
		return principal
	}

	return verified
}

func withActor(ctx context.Context, principal authdomain.Principal) context.Context {
	if principal.IsAccount() {
		return actorctx.WithActorID(ctx, principal.AccountID)
	}
	return ctx
}

func extractBearer(v string) string {
	const prefix = "Bearer "
	if len(v) < len(prefix)+1 || !strings.EqualFold(v[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(v[len(prefix):])
}
