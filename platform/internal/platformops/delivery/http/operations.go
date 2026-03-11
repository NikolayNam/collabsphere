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

var listAutoGrantRulesOp = huma.Operation{
	OperationID: "platform-access-list-auto-grant-rules",
	Method:      "GET",
	Path:        "/platform/access/auto-grant-rules",
	Tags:        []string{"Platform / Access"},
	Summary:     "List platform auto-grant rules",
	Description: "Returns effective auto-grant rules used during first successful OIDC login. Bootstrap rules loaded from YAML are listed with source `bootstrap_config` and are read-only; API-managed rules use source `database`.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createAutoGrantRuleOp = huma.Operation{
	OperationID: "platform-access-create-auto-grant-rule",
	Method:      "POST",
	Path:        "/platform/access/auto-grant-rules",
	Tags:        []string{"Platform / Access"},
	Summary:     "Create a platform auto-grant rule",
	Description: "Adds a database-backed auto-grant rule so future successful OIDC logins automatically receive the target platform role. Email rules apply only when ZITADEL returns `email_verified=true`; subject rules match regardless of email verification.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var deleteAutoGrantRuleOp = huma.Operation{
	OperationID: "platform-access-delete-auto-grant-rule",
	Method:      "DELETE",
	Path:        "/platform/access/auto-grant-rules/{ruleId}",
	Tags:        []string{"Platform / Access"},
	Summary:     "Delete a platform auto-grant rule",
	Description: "Deletes a database-backed auto-grant rule. Bootstrap rules loaded from YAML are not deletable through the API because they are part of the backend's bootstrap configuration.\n\n" + platformControlPlaneCaution,
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
