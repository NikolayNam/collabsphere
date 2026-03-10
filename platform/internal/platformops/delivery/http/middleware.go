package http

import (
	"context"
	"net/http"
	"strings"

	platformapp "github.com/NikolayNam/collabsphere/internal/platformops/application"
	platformdomain "github.com/NikolayNam/collabsphere/internal/platformops/domain"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

type platformAccessCtxKey struct{}

func withPlatformAccess(ctx context.Context, access *platformdomain.Access) context.Context {
	return context.WithValue(ctx, platformAccessCtxKey{}, access)
}

func platformAccessFromContext(ctx context.Context) *platformdomain.Access {
	value := ctx.Value(platformAccessCtxKey{})
	access, _ := value.(*platformdomain.Access)
	return access
}

func (h *Handler) roleGuard(api huma.API, allowed ...platformdomain.Role) func(huma.Context, func(huma.Context)) {
	return func(hctx huma.Context, next func(huma.Context)) {
		principal := authmw.PrincipalFromContext(hctx.Context())
		action := strings.TrimSpace(hctx.Operation().OperationID)
		targetID := strings.TrimSpace(hctx.Operation().Path)
		if !principal.IsAccount() {
			h.svc.RecordDeniedAudit(hctx.Context(), platformapp.AuditDeniedCmd{
				Action:     action,
				TargetType: "operation",
				TargetID:   targetID,
				Summary:    "authentication required",
			})
			_ = huma.WriteErr(api, hctx, http.StatusUnauthorized, "Authentication required")
			return
		}

		access, err := h.svc.ResolveAccess(hctx.Context(), principal.AccountID)
		if err != nil {
			_ = huma.WriteErr(api, hctx, http.StatusInternalServerError, "Unable to resolve platform access", err)
			return
		}
		ctx := withPlatformAccess(hctx.Context(), access)
		if !access.HasAnyRole(allowed...) {
			actorID := principal.AccountID
			h.svc.RecordDeniedAudit(ctx, platformapp.AuditDeniedCmd{
				ActorAccountID: &actorID,
				ActorRoles:     access.EffectiveRoles,
				ActorBootstrap: access.BootstrapAdmin,
				Action:         action,
				TargetType:     "operation",
				TargetID:       targetID,
				Summary:        "required roles: " + strings.Join(platformdomain.RoleStrings(allowed), ", "),
			})
			_ = huma.WriteErr(api, huma.WithContext(hctx, ctx), http.StatusForbidden, "Platform access denied")
			return
		}

		next(huma.WithContext(hctx, ctx))
	}
}
