package http

import (
	"context"
	"strings"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

type AccessTokenVerifier interface {
	VerifyAccessToken(ctx context.Context, token string) (authdomain.Principal, error)
}

func authPrincipalMiddleware(verifier AccessTokenVerifier) func(huma.Context, func(huma.Context)) {
	return func(hctx huma.Context, next func(huma.Context)) {
		principal := authdomain.AnonymousPrincipal()

		if verifier != nil {
			if token := extractBearer(hctx.Header("Authorization")); token != "" {
				if verified, err := verifier.VerifyAccessToken(hctx.Context(), token); err == nil {
					principal = verified
				}
			}
		}

		ctx := authmw.WithPrincipal(hctx.Context(), principal)
		next(huma.WithContext(hctx, ctx))
	}
}

func extractBearer(value string) string {
	value = strings.TrimSpace(value)
	if len(value) < len("Bearer ")+1 {
		return ""
	}
	if !strings.EqualFold(value[:7], "Bearer ") {
		return ""
	}
	return strings.TrimSpace(value[7:])
}
