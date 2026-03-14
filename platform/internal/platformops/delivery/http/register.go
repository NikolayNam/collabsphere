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
	reviewOrAdmin := huma.Middlewares{authmw.HumaAuthOptional(verifier), h.roleGuard(api, platformdomain.RolePlatformAdmin, platformdomain.RoleReviewOperator)}

	getMyAccess := getMyAccessOp
	getMyAccess.Middlewares = anyPlatformRole
	huma.Register(api, getMyAccess, h.GetMyAccess)

	getAccountRoles := getAccountRolesOp
	getAccountRoles.Middlewares = adminOnly
	huma.Register(api, getAccountRoles, h.GetAccountRoles)

	replaceAccountRoles := replaceAccountRolesOp
	replaceAccountRoles.Middlewares = adminOnly
	huma.Register(api, replaceAccountRoles, h.ReplaceAccountRoles)

	listAutoGrantRules := listAutoGrantRulesOp
	listAutoGrantRules.Middlewares = adminOnly
	huma.Register(api, listAutoGrantRules, h.ListAutoGrantRules)

	createAutoGrantRule := createAutoGrantRuleOp
	createAutoGrantRule.Middlewares = adminOnly
	huma.Register(api, createAutoGrantRule, h.CreateAutoGrantRule)

	deleteAutoGrantRule := deleteAutoGrantRuleOp
	deleteAutoGrantRule.Middlewares = adminOnly
	huma.Register(api, deleteAutoGrantRule, h.DeleteAutoGrantRule)

	dashboardSummary := dashboardSummaryOp
	dashboardSummary.Middlewares = anyPlatformRole
	huma.Register(api, dashboardSummary, h.GetDashboardSummary)

	listUploads := listUploadsOp
	listUploads.Middlewares = supportOrAdmin
	huma.Register(api, listUploads, h.ListUploads)

	listReviews := listOrganizationReviewsOp
	listReviews.Middlewares = anyPlatformRole
	huma.Register(api, listReviews, h.ListOrganizationReviews)

	getReview := getOrganizationReviewOp
	getReview.Middlewares = anyPlatformRole
	huma.Register(api, getReview, h.GetOrganizationReview)

	transitionReview := transitionCooperationApplicationReviewOp
	transitionReview.Middlewares = reviewOrAdmin
	huma.Register(api, transitionReview, h.TransitionCooperationApplicationReview)

	transitionLegalDocument := transitionLegalDocumentReviewOp
	transitionLegalDocument.Middlewares = reviewOrAdmin
	huma.Register(api, transitionLegalDocument, h.TransitionLegalDocumentReview)

	listKYCReviews := listKYCReviewsOp
	listKYCReviews.Middlewares = reviewOrAdmin
	huma.Register(api, listKYCReviews, h.ListKYCReviews)

	getKYCReview := getKYCReviewOp
	getKYCReview.Middlewares = reviewOrAdmin
	huma.Register(api, getKYCReview, h.GetKYCReview)

	decideKYCReview := decideKYCReviewOp
	decideKYCReview.Middlewares = reviewOrAdmin
	huma.Register(api, decideKYCReview, h.DecideKYCReview)

	decideKYCDocument := decideKYCDocumentOp
	decideKYCDocument.Middlewares = reviewOrAdmin
	huma.Register(api, decideKYCDocument, h.DecideKYCDocument)

	listKYCLevels := listKYCLevelsOp
	listKYCLevels.Middlewares = anyPlatformRole
	huma.Register(api, listKYCLevels, h.ListKYCLevels)

	createKYCLevel := createKYCLevelOp
	createKYCLevel.Middlewares = adminOnly
	huma.Register(api, createKYCLevel, h.CreateKYCLevel)

	updateKYCLevel := updateKYCLevelOp
	updateKYCLevel.Middlewares = adminOnly
	huma.Register(api, updateKYCLevel, h.UpdateKYCLevel)

	deleteKYCLevel := deleteKYCLevelOp
	deleteKYCLevel.Middlewares = adminOnly
	huma.Register(api, deleteKYCLevel, h.DeleteKYCLevel)

	issueKYCLevel := issueKYCLevelOp
	issueKYCLevel.Middlewares = reviewOrAdmin
	huma.Register(api, issueKYCLevel, h.IssueKYCLevel)

	forceVerify := forceVerifyUserEmailOp
	forceVerify.Middlewares = adminOnly
	if !h.svc.ZitadelAdminEnabled() {
		forceVerify.Description += "\n\n> [!CAUTION]\n> This control-plane action is disabled until `AUTH_ZITADEL_ADMIN_TOKEN` or `AUTH_ZITADEL_ADMIN_TOKEN_FILE` is configured on the backend."
	}
	huma.Register(api, forceVerify, h.ForceVerifyUserEmail)
}
