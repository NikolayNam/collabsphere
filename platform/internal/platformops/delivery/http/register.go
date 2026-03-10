package http

import (
	platformdomain "github.com/NikolayNam/collabsphere/internal/platformops/domain"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	anyPlatformRole := huma.Middlewares{authmw.HumaAuthOptional(verifier), h.roleGuard(api, platformdomain.RolePlatformAdmin, platformdomain.RoleSupportOperator, platformdomain.RoleReviewOperator)}
	adminOnly := huma.Middlewares{authmw.HumaAuthOptional(verifier), h.roleGuard(api, platformdomain.RolePlatformAdmin)}
	supportOrAdmin := huma.Middlewares{authmw.HumaAuthOptional(verifier), h.roleGuard(api, platformdomain.RolePlatformAdmin, platformdomain.RoleSupportOperator)}

	getMyAccess := getMyAccessOp
	getMyAccess.Middlewares = anyPlatformRole
	huma.Register(api, getMyAccess, h.GetMyAccess)

	getAccountRoles := getAccountRolesOp
	getAccountRoles.Middlewares = adminOnly
	huma.Register(api, getAccountRoles, h.GetAccountRoles)

	replaceAccountRoles := replaceAccountRolesOp
	replaceAccountRoles.Middlewares = adminOnly
	huma.Register(api, replaceAccountRoles, h.ReplaceAccountRoles)

	dashboardSummary := dashboardSummaryOp
	dashboardSummary.Middlewares = anyPlatformRole
	huma.Register(api, dashboardSummary, h.GetDashboardSummary)

	listUploads := listUploadsOp
	listUploads.Middlewares = supportOrAdmin
	huma.Register(api, listUploads, h.ListUploads)

	forceVerify := forceVerifyUserEmailOp
	forceVerify.Middlewares = adminOnly
	if !h.svc.ZitadelAdminEnabled() {
		forceVerify.Description += "\n\n> [!CAUTION]\n> This control-plane action is disabled until `AUTH_ZITADEL_ADMIN_TOKEN` or `AUTH_ZITADEL_ADMIN_TOKEN_FILE` is configured on the backend."
	}
	huma.Register(api, forceVerify, h.ForceVerifyUserEmail)

	forceVerifyAlias := forceVerifyUserEmailAliasOp
	forceVerifyAlias.Middlewares = adminOnly
	if !h.svc.ZitadelAdminEnabled() {
		forceVerifyAlias.Description += "\n\n> [!CAUTION]\n> This deprecated alias is disabled until `AUTH_ZITADEL_ADMIN_TOKEN` or `AUTH_ZITADEL_ADMIN_TOKEN_FILE` is configured on the backend."
	}
	huma.Register(api, forceVerifyAlias, h.ForceVerifyUserEmail)
}
