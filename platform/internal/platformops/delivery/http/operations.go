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

var listOrganizationReviewsOp = huma.Operation{
	OperationID: "platform-list-organization-reviews",
	Method:      "GET",
	Path:        "/platform/reviews/organizations/cooperation-applications",
	Tags:        []string{"Platform / Reviews"},
	Summary:     "List organization cooperation review queue",
	Description: "Returns the global cooperation application review queue for platform reviewers, admins, and support operators.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getOrganizationReviewOp = huma.Operation{
	OperationID: "platform-get-organization-review",
	Method:      "GET",
	Path:        "/platform/reviews/organizations/{organizationId}",
	Tags:        []string{"Platform / Reviews"},
	Summary:     "Get organization review card",
	Description: "Returns the control-plane review card for one organization, including cooperation application data, active domains, and legal document summaries.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var transitionCooperationApplicationReviewOp = huma.Operation{
	OperationID: "platform-transition-cooperation-application-review",
	Method:      "POST",
	Path:        "/platform/reviews/organizations/{organizationId}/cooperation-application/transition",
	Tags:        []string{"Platform / Reviews"},
	Summary:     "Transition cooperation application review status",
	Description: "Moves a cooperation application between allowed control-plane review states and records the acting reviewer in platform audit logs.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var transitionLegalDocumentReviewOp = huma.Operation{
	OperationID: "platform-transition-legal-document-review",
	Method:      "POST",
	Path:        "/platform/reviews/organizations/{organizationId}/legal-documents/{documentId}/transition",
	Tags:        []string{"Platform / Reviews"},
	Summary:     "Transition legal document review status",
	Description: "Approves or rejects a legal document from the control-plane review card and records the acting reviewer in platform audit logs.\n\n" + platformControlPlaneCaution,
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

var listKYCReviewsOp = huma.Operation{
	OperationID: "platform-list-kyc-reviews",
	Method:      "GET",
	Path:        "/platform/kyc/reviews",
	Tags:        []string{"Platform / Reviews"},
	Summary:     "List KYC reviews",
	Description: "Returns account and organization KYC reviews for platform reviewers.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getKYCReviewOp = huma.Operation{
	OperationID: "platform-get-kyc-review",
	Method:      "GET",
	Path:        "/platform/kyc/reviews/{reviewId}",
	Tags:        []string{"Platform / Reviews"},
	Summary:     "Get KYC review details",
	Description: "Returns a single KYC review card for account or organization scope.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var decideKYCReviewOp = huma.Operation{
	OperationID: "platform-decide-kyc-review",
	Method:      "POST",
	Path:        "/platform/kyc/reviews/{reviewId}/decision",
	Tags:        []string{"Platform / Reviews"},
	Summary:     "Apply KYC review decision",
	Description: "Applies approve/reject/request_info decision to account or organization KYC profile.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var decideKYCDocumentOp = huma.Operation{
	OperationID: "platform-decide-kyc-document",
	Method:      "POST",
	Path:        "/platform/kyc/reviews/{reviewId}/documents/{documentId}/decision",
	Tags:        []string{"Platform / Reviews"},
	Summary:     "Apply KYC document decision",
	Description: "Applies approve/reject/request_info decision to a single KYC document and recalculates aggregate review status.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listKYCLevelsOp = huma.Operation{
	OperationID: "platform-list-kyc-levels",
	Method:      "GET",
	Path:        "/platform/kyc/levels",
	Tags:        []string{"Platform / Reviews"},
	Summary:     "List configurable KYC levels",
	Description: "Returns configured KYC levels and document requirements.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createKYCLevelOp = huma.Operation{
	OperationID: "platform-create-kyc-level",
	Method:      "POST",
	Path:        "/platform/kyc/levels",
	Tags:        []string{"Platform / Reviews"},
	Summary:     "Create KYC level",
	Description: "Creates configurable KYC level with document requirements.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var updateKYCLevelOp = huma.Operation{
	OperationID: "platform-update-kyc-level",
	Method:      "PUT",
	Path:        "/platform/kyc/levels/{levelId}",
	Tags:        []string{"Platform / Reviews"},
	Summary:     "Update KYC level",
	Description: "Updates configurable KYC level and replaces document requirements.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var deleteKYCLevelOp = huma.Operation{
	OperationID: "platform-delete-kyc-level",
	Method:      "DELETE",
	Path:        "/platform/kyc/levels/{levelId}",
	Tags:        []string{"Platform / Reviews"},
	Summary:     "Delete KYC level",
	Description: "Deletes configurable KYC level and all attached requirements.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var issueKYCLevelOp = huma.Operation{
	OperationID: "platform-issue-kyc-level",
	Method:      "POST",
	Path:        "/platform/kyc/reviews/{reviewId}/issue-level",
	Tags:        []string{"Platform / Reviews"},
	Summary:     "Issue KYC level by verified documents",
	Description: "Evaluates configured requirements against verified documents and assigns the highest matching KYC level.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listAttachmentLimitsOp = huma.Operation{
	OperationID: "platform-list-attachment-limits",
	Method:      "GET",
	Path:        "/platform/attachment-limits",
	Tags:        []string{"Platform / Attachment Limits"},
	Summary:     "List attachment limits",
	Description: "Returns attachment limits with optional filters by scope type and scope id. Resolution order: account > organization > platform.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getPlatformAttachmentLimitOp = huma.Operation{
	OperationID: "platform-get-platform-attachment-limit",
	Method:      "GET",
	Path:        "/platform/attachment-limits/platform",
	Tags:        []string{"Platform / Attachment Limits"},
	Summary:     "Get platform default attachment limit",
	Description: "Returns the platform-wide default attachment limits used when no org/account override exists.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var upsertPlatformAttachmentLimitOp = huma.Operation{
	OperationID: "platform-upsert-platform-attachment-limit",
	Method:      "PUT",
	Path:        "/platform/attachment-limits/platform",
	Tags:        []string{"Platform / Attachment Limits"},
	Summary:     "Update platform default attachment limit",
	Description: "Creates or updates the platform-wide default attachment limits.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getOrganizationAttachmentLimitOp = huma.Operation{
	OperationID: "platform-get-organization-attachment-limit",
	Method:      "GET",
	Path:        "/platform/attachment-limits/organizations/{organizationId}",
	Tags:        []string{"Platform / Attachment Limits"},
	Summary:     "Get organization attachment limit",
	Description: "Returns organization-specific attachment limits override.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var upsertOrganizationAttachmentLimitOp = huma.Operation{
	OperationID: "platform-upsert-organization-attachment-limit",
	Method:      "PUT",
	Path:        "/platform/attachment-limits/organizations/{organizationId}",
	Tags:        []string{"Platform / Attachment Limits"},
	Summary:     "Create or update organization attachment limit",
	Description: "Creates or updates organization-specific attachment limits override.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var deleteOrganizationAttachmentLimitOp = huma.Operation{
	OperationID: "platform-delete-organization-attachment-limit",
	Method:      "DELETE",
	Path:        "/platform/attachment-limits/organizations/{organizationId}",
	Tags:        []string{"Platform / Attachment Limits"},
	Summary:     "Delete organization attachment limit",
	Description: "Removes organization-specific override; organization falls back to platform default.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getAccountAttachmentLimitOp = huma.Operation{
	OperationID: "platform-get-account-attachment-limit",
	Method:      "GET",
	Path:        "/platform/attachment-limits/accounts/{accountId}",
	Tags:        []string{"Platform / Attachment Limits"},
	Summary:     "Get account attachment limit",
	Description: "Returns account-specific attachment limits override.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var upsertAccountAttachmentLimitOp = huma.Operation{
	OperationID: "platform-upsert-account-attachment-limit",
	Method:      "PUT",
	Path:        "/platform/attachment-limits/accounts/{accountId}",
	Tags:        []string{"Platform / Attachment Limits"},
	Summary:     "Create or update account attachment limit",
	Description: "Creates or updates account-specific attachment limits override.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var deleteAccountAttachmentLimitOp = huma.Operation{
	OperationID: "platform-delete-account-attachment-limit",
	Method:      "DELETE",
	Path:        "/platform/attachment-limits/accounts/{accountId}",
	Tags:        []string{"Platform / Attachment Limits"},
	Summary:     "Delete account attachment limit",
	Description: "Removes account-specific override; account falls back to organization or platform default.\n\n" + platformControlPlaneCaution,
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
