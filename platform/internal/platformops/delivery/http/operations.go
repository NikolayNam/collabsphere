package http

import "github.com/danielgtaylor/huma/v2"

const platformControlPlaneCaution = "> [!CAUTION]\n> Internal control-plane endpoint. Not for tenant organization admins."

var getMyAccessOp = huma.Operation{
	OperationID: "platform-access-me",
	Method:      "GET",
	Path:        "/platform/access/me",
	Tags:        []string{"Platform / Access"},
	Summary:     "Get my platform access",
	Description: "Returns the caller's effective control-plane roles and whether they are receiving emergency bootstrap admin access.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getAccountRolesOp = huma.Operation{
	OperationID: "platform-access-account-roles",
	Method:      "GET",
	Path:        "/platform/access/accounts/{accountId}/roles",
	Tags:        []string{"Platform / Access"},
	Summary:     "Get platform roles for an account",
	Description: "Resolves stored and effective global platform roles for a local account.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var replaceAccountRolesOp = huma.Operation{
	OperationID: "platform-access-replace-account-roles",
	Method:      "PUT",
	Path:        "/platform/access/accounts/{accountId}/roles",
	Tags:        []string{"Platform / Access"},
	Summary:     "Replace platform roles for an account",
	Description: "Replaces the stored global platform roles for a local account. Bootstrap admin access from config is not removed by this route.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var dashboardSummaryOp = huma.Operation{
	OperationID: "platform-dashboard-summary",
	Method:      "GET",
	Path:        "/platform/dashboards/summary",
	Tags:        []string{"Platform / Dashboards"},
	Summary:     "Get platform dashboard summary",
	Description: "Returns global counts for accounts, organizations, uploads, and cooperation review states.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listUploadsOp = huma.Operation{
	OperationID: "platform-list-uploads",
	Method:      "GET",
	Path:        "/platform/uploads",
	Tags:        []string{"Platform / Uploads"},
	Summary:     "List upload queue items",
	Description: "Returns tracked upload sessions across the whole platform so support operators can inspect pending, ready, or failed upload flows.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var forceVerifyUserEmailOp = huma.Operation{
	OperationID: "platform-force-verify-zitadel-user-email",
	Method:      "POST",
	Path:        "/platform/users/{userId}/email/force-verify",
	Tags:        []string{"Platform / Users"},
	Summary:     "Force-verify a ZITADEL user email",
	Description: "Uses the backend's server-side ZITADEL admin token to verify an existing user's email. The backend first requests a verification code from ZITADEL and then immediately verifies the email with that code.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var forceVerifyUserEmailAliasOp = huma.Operation{
	OperationID: "admin-force-verify-zitadel-user-email-alias",
	Method:      "POST",
	Path:        "/admin/zitadel/users/{userId}/email/force-verify",
	Tags:        []string{"Platform / Users"},
	Summary:     "Force-verify a ZITADEL user email (deprecated alias)",
	Description: "Deprecated alias for `POST /platform/users/{userId}/email/force-verify`.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
	Deprecated:  true,
}
