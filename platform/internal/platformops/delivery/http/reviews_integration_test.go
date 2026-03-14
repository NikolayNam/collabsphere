//go:build integration

package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	accpg "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	platformapp "github.com/NikolayNam/collabsphere/internal/platformops/application"
	platformdomain "github.com/NikolayNam/collabsphere/internal/platformops/domain"
	platformpg "github.com/NikolayNam/collabsphere/internal/platformops/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/bootstrap"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	"github.com/NikolayNam/collabsphere/internal/testutil/postgresitest"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestListOrganizationReviewsIntegrationReturnsQueueForSupportOperator(t *testing.T) {
	env := newPlatformReviewsIntegrationEnv(t)
	_, token := env.createAuthenticatedPlatformAccount(t, "support-review@example.com", platformdomain.RoleSupportOperator)

	targetOrgID := env.insertOrganization(t, "Acme Foods", "acme-foods")
	env.insertCooperationApplication(t, platformReviewApplicationSeed{
		OrganizationID:    targetOrgID,
		Status:            "submitted",
		CompanyName:       stringPtr("Acme Wholesale"),
		ConfirmationEmail: stringPtr("buyer@acme.test"),
	})

	ignoredOrgID := env.insertOrganization(t, "Ignored Org", "ignored-org")
	env.insertCooperationApplication(t, platformReviewApplicationSeed{
		OrganizationID: ignoredOrgID,
		Status:         "approved",
	})

	resp := env.get(t, "/platform/reviews/organizations/cooperation-applications?status=submitted&q=acme", token)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Total int `json:"total"`
		Items []struct {
			OrganizationID    uuid.UUID `json:"organizationId"`
			OrganizationName  string    `json:"organizationName"`
			OrganizationSlug  string    `json:"organizationSlug"`
			CooperationStatus string    `json:"cooperationStatus"`
			CompanyName       *string   `json:"companyName"`
		} `json:"items"`
	}
	decodePlatformReviewsJSON(t, resp.Body, &body)

	if body.Total != 1 {
		t.Fatalf("total = %d, want 1", body.Total)
	}
	if len(body.Items) != 1 {
		t.Fatalf("items len = %d, want 1", len(body.Items))
	}
	item := body.Items[0]
	if item.OrganizationID != targetOrgID {
		t.Fatalf("organizationId = %s, want %s", item.OrganizationID, targetOrgID)
	}
	if item.OrganizationName != "Acme Foods" {
		t.Fatalf("organizationName = %q, want Acme Foods", item.OrganizationName)
	}
	if item.OrganizationSlug != "acme-foods" {
		t.Fatalf("organizationSlug = %q, want acme-foods", item.OrganizationSlug)
	}
	if item.CooperationStatus != "submitted" {
		t.Fatalf("cooperationStatus = %q, want submitted", item.CooperationStatus)
	}
	if item.CompanyName == nil || *item.CompanyName != "Acme Wholesale" {
		t.Fatalf("companyName = %v, want Acme Wholesale", item.CompanyName)
	}
}

func TestGetOrganizationReviewIntegrationReturnsAggregate(t *testing.T) {
	env := newPlatformReviewsIntegrationEnv(t)
	uploaderID, token := env.createAuthenticatedPlatformAccount(t, "review-reader@example.com", platformdomain.RoleReviewOperator)

	organizationID := env.insertOrganization(t, "Review Org", "review-org")
	env.insertDomain(t, organizationID, "review-org.collabsphere.ru", "subdomain", true, true)
	priceListObjectID := env.insertStorageObject(t, organizationID, "integration-bucket", "organizations/cooperation/review-org/price-list.csv", "price-list.csv", "text/csv", 64)
	env.insertCooperationApplication(t, platformReviewApplicationSeed{
		OrganizationID:        organizationID,
		Status:                "under_review",
		CompanyName:           stringPtr("Review Trading"),
		ConfirmationEmail:     stringPtr("hello@review.test"),
		RepresentedCategories: stringPtr("Food"),
		MinimumOrderAmount:    stringPtr("1000"),
		DeliveryGeography:     stringPtr("Moscow"),
		SalesChannels:         []string{"Retail"},
		PriceListObjectID:     &priceListObjectID,
		ContactFirstName:      stringPtr("Ivan"),
		ContactLastName:       stringPtr("Petrov"),
		ContactJobTitle:       stringPtr("Manager"),
		ContactEmail:          stringPtr("sales@review.test"),
		ContactPhone:          stringPtr("+79990000000"),
		ReviewNote:            stringPtr("Initial review started"),
		ReviewerAccountID:     &uploaderID,
		SubmittedAt:           timePtr(time.Now().UTC().Add(-30 * time.Minute)),
	})
	objectID := env.insertStorageObject(t, organizationID, "integration-bucket", "organizations/legal/review-org/vat.pdf", "vat.pdf", "application/pdf", 128)
	documentID := env.insertLegalDocument(t, organizationID, objectID, uploaderID, platformReviewLegalDocumentSeed{
		DocumentType: "tax_certificate",
		Status:       "pending",
		Title:        "VAT Certificate",
	})
	env.insertLegalDocumentAnalysis(t, organizationID, documentID, platformReviewAnalysisSeed{
		Status:               "completed",
		Provider:             "openai",
		Summary:              stringPtr("Document looks valid"),
		ExtractedFieldsJSON:  `{"companyName":"Review Trading"}`,
		DetectedDocumentType: stringPtr("tax_certificate"),
		ConfidenceScore:      floatPtr(0.91),
	})

	resp := env.get(t, "/platform/reviews/organizations/"+organizationID.String(), token)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Organization struct {
			ID   uuid.UUID `json:"id"`
			Name string    `json:"name"`
			Slug string    `json:"slug"`
		} `json:"organization"`
		Domains []struct {
			Hostname   string `json:"hostname"`
			IsPrimary  bool   `json:"isPrimary"`
			IsVerified bool   `json:"isVerified"`
		} `json:"domains"`
		CooperationApplication *struct {
			Status      string  `json:"status"`
			CompanyName *string `json:"companyName"`
			ReviewNote  *string `json:"reviewNote"`
		} `json:"cooperationApplication"`
		LegalDocuments []struct {
			Title    string `json:"title"`
			Status   string `json:"status"`
			Analysis *struct {
				Status  string  `json:"status"`
				Summary *string `json:"summary"`
			} `json:"analysis"`
			Verification *struct {
				Verdict       string   `json:"verdict"`
				MissingFields []string `json:"missingFields"`
			} `json:"verification"`
		} `json:"legalDocuments"`
		KYC *struct {
			Status              string `json:"status"`
			PendingVerification []struct {
				Code string `json:"code"`
			} `json:"pendingVerification"`
		} `json:"kyc"`
	}
	decodePlatformReviewsJSON(t, resp.Body, &body)

	if body.Organization.ID != organizationID {
		t.Fatalf("organization.id = %s, want %s", body.Organization.ID, organizationID)
	}
	if body.Organization.Name != "Review Org" {
		t.Fatalf("organization.name = %q, want Review Org", body.Organization.Name)
	}
	if body.Organization.Slug != "review-org" {
		t.Fatalf("organization.slug = %q, want review-org", body.Organization.Slug)
	}
	if len(body.Domains) != 1 {
		t.Fatalf("domains len = %d, want 1", len(body.Domains))
	}
	if body.Domains[0].Hostname != "review-org.collabsphere.ru" {
		t.Fatalf("domain hostname = %q, want review-org.collabsphere.ru", body.Domains[0].Hostname)
	}
	if !body.Domains[0].IsPrimary || !body.Domains[0].IsVerified {
		t.Fatalf("domain = %+v, want primary verified domain", body.Domains[0])
	}
	if body.CooperationApplication == nil {
		t.Fatal("cooperationApplication = nil, want aggregate")
	}
	if body.CooperationApplication.Status != "under_review" {
		t.Fatalf("cooperationApplication.status = %q, want under_review", body.CooperationApplication.Status)
	}
	if body.CooperationApplication.CompanyName == nil || *body.CooperationApplication.CompanyName != "Review Trading" {
		t.Fatalf("cooperationApplication.companyName = %v, want Review Trading", body.CooperationApplication.CompanyName)
	}
	if len(body.LegalDocuments) != 1 {
		t.Fatalf("legalDocuments len = %d, want 1", len(body.LegalDocuments))
	}
	if body.LegalDocuments[0].Title != "VAT Certificate" {
		t.Fatalf("legalDocuments[0].title = %q, want VAT Certificate", body.LegalDocuments[0].Title)
	}
	if body.LegalDocuments[0].Analysis == nil {
		t.Fatal("legalDocuments[0].analysis = nil, want summary")
	}
	if body.LegalDocuments[0].Analysis.Status != "completed" {
		t.Fatalf("legalDocuments[0].analysis.status = %q, want completed", body.LegalDocuments[0].Analysis.Status)
	}
	if body.LegalDocuments[0].Analysis.Summary == nil || *body.LegalDocuments[0].Analysis.Summary != "Document looks valid" {
		t.Fatalf("legalDocuments[0].analysis.summary = %v, want Document looks valid", body.LegalDocuments[0].Analysis.Summary)
	}
	if body.LegalDocuments[0].Verification == nil {
		t.Fatal("legalDocuments[0].verification = nil, want machine verdict")
	}
	if body.LegalDocuments[0].Verification.Verdict != "approved" {
		t.Fatalf("legalDocuments[0].verification.verdict = %q, want approved", body.LegalDocuments[0].Verification.Verdict)
	}
	if len(body.LegalDocuments[0].Verification.MissingFields) != 0 {
		t.Fatalf("legalDocuments[0].verification.missingFields = %v, want empty", body.LegalDocuments[0].Verification.MissingFields)
	}
	if body.KYC == nil {
		t.Fatal("kyc = nil, want aggregated KYC snapshot")
	}
	if body.KYC.Status != "pending_verification" {
		t.Fatalf("kyc.status = %q, want pending_verification", body.KYC.Status)
	}
	if len(body.KYC.PendingVerification) != 2 {
		t.Fatalf("kyc.pendingVerification len = %d, want 2", len(body.KYC.PendingVerification))
	}
}

func TestTransitionCooperationApplicationReviewIntegrationUpdatesStateAndWritesAudit(t *testing.T) {
	env := newPlatformReviewsIntegrationEnv(t)
	reviewerID, token := env.createAuthenticatedPlatformAccount(t, "reviewer-transition@example.com", platformdomain.RoleReviewOperator)

	organizationID := env.insertOrganization(t, "Needs Info Org", "needs-info-org")
	env.insertCooperationApplication(t, platformReviewApplicationSeed{
		OrganizationID: organizationID,
		Status:         "under_review",
	})

	resp := env.postJSON(t, "/platform/reviews/organizations/"+organizationID.String()+"/cooperation-application/transition", map[string]any{
		"targetStatus": "needs_info",
		"reviewNote":   "Need VAT certificate",
	}, token)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		OrganizationID    uuid.UUID  `json:"organizationId"`
		Status            string     `json:"status"`
		ReviewNote        *string    `json:"reviewNote"`
		ReviewerAccountID *uuid.UUID `json:"reviewerAccountId"`
		ReviewedAt        *time.Time `json:"reviewedAt"`
	}
	decodePlatformReviewsJSON(t, resp.Body, &body)

	if body.OrganizationID != organizationID {
		t.Fatalf("organizationId = %s, want %s", body.OrganizationID, organizationID)
	}
	if body.Status != "needs_info" {
		t.Fatalf("status = %q, want needs_info", body.Status)
	}
	if body.ReviewNote == nil || *body.ReviewNote != "Need VAT certificate" {
		t.Fatalf("reviewNote = %v, want Need VAT certificate", body.ReviewNote)
	}
	if body.ReviewerAccountID == nil || *body.ReviewerAccountID != reviewerID {
		t.Fatalf("reviewerAccountId = %v, want %s", body.ReviewerAccountID, reviewerID)
	}
	if body.ReviewedAt == nil || body.ReviewedAt.IsZero() {
		t.Fatal("reviewedAt = nil/zero, want timestamp")
	}

	var (
		storedStatus     string
		storedNote       string
		storedReviewerID uuid.UUID
		storedReviewedAt time.Time
	)
	if err := env.queryRowContext(t, `
		SELECT status, review_note, reviewer_account_id, reviewed_at
		FROM org.cooperation_applications
		WHERE organization_id = $1
	`, organizationID).Scan(&storedStatus, &storedNote, &storedReviewerID, &storedReviewedAt); err != nil {
		t.Fatalf("query cooperation_applications: %v", err)
	}
	if storedStatus != "needs_info" {
		t.Fatalf("db status = %q, want needs_info", storedStatus)
	}
	if storedNote != "Need VAT certificate" {
		t.Fatalf("db review_note = %q, want Need VAT certificate", storedNote)
	}
	if storedReviewerID != reviewerID {
		t.Fatalf("db reviewer_account_id = %s, want %s", storedReviewerID, reviewerID)
	}
	if storedReviewedAt.IsZero() {
		t.Fatal("db reviewed_at = zero, want timestamp")
	}

	var (
		auditAction  string
		auditStatus  string
		auditTarget  string
		auditSummary string
	)
	if err := env.queryRowContext(t, `
		SELECT action, status, target_id, summary
		FROM iam.platform_audit_events
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(&auditAction, &auditStatus, &auditTarget, &auditSummary); err != nil {
		t.Fatalf("query platform audit: %v", err)
	}
	if auditAction != "platform.organization.review.transition" {
		t.Fatalf("audit action = %q, want platform.organization.review.transition", auditAction)
	}
	if auditStatus != "success" {
		t.Fatalf("audit status = %q, want success", auditStatus)
	}
	if auditTarget != organizationID.String() {
		t.Fatalf("audit target_id = %q, want %s", auditTarget, organizationID)
	}
	if auditSummary != "from=under_review to=needs_info" {
		t.Fatalf("audit summary = %q, want from=under_review to=needs_info", auditSummary)
	}
}

func TestTransitionCooperationApplicationReviewIntegrationRejectsSupportOperatorAndWritesDeniedAudit(t *testing.T) {
	env := newPlatformReviewsIntegrationEnv(t)
	supportID, token := env.createAuthenticatedPlatformAccount(t, "support-transition@example.com", platformdomain.RoleSupportOperator)

	organizationID := env.insertOrganization(t, "Denied Org", "denied-org")
	env.insertCooperationApplication(t, platformReviewApplicationSeed{
		OrganizationID: organizationID,
		Status:         "under_review",
	})

	resp := env.postJSON(t, "/platform/reviews/organizations/"+organizationID.String()+"/cooperation-application/transition", map[string]any{
		"targetStatus": "approved",
	}, token)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}

	var (
		actorAccountID uuid.UUID
		action         string
		targetType     string
		status         string
		summary        string
	)
	if err := env.queryRowContext(t, `
		SELECT actor_account_id, action, target_type, status, summary
		FROM iam.platform_audit_events
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(&actorAccountID, &action, &targetType, &status, &summary); err != nil {
		t.Fatalf("query denied audit: %v", err)
	}
	if actorAccountID != supportID {
		t.Fatalf("actor_account_id = %s, want %s", actorAccountID, supportID)
	}
	if action != "platform-transition-cooperation-application-review" {
		t.Fatalf("action = %q, want platform-transition-cooperation-application-review", action)
	}
	if targetType != "operation" {
		t.Fatalf("targetType = %q, want operation", targetType)
	}
	if status != "denied" {
		t.Fatalf("status = %q, want denied", status)
	}
	if !strings.Contains(summary, "platform_admin, review_operator") {
		t.Fatalf("summary = %q, want required roles message", summary)
	}
}

func TestTransitionLegalDocumentReviewIntegrationUpdatesStateAndWritesAudit(t *testing.T) {
	env := newPlatformReviewsIntegrationEnv(t)
	reviewerID, token := env.createAuthenticatedPlatformAccount(t, "legal-reviewer@example.com", platformdomain.RoleReviewOperator)
	uploaderID, _ := env.createAuthenticatedPlatformAccount(t, "legal-uploader@example.com", platformdomain.RoleSupportOperator)

	organizationID := env.insertOrganization(t, "Docs Org", "docs-org")
	objectID := env.insertStorageObject(t, organizationID, "integration-bucket", "organizations/legal/docs-org/charter.pdf", "charter.pdf", "application/pdf", 256)
	documentID := env.insertLegalDocument(t, organizationID, objectID, uploaderID, platformReviewLegalDocumentSeed{
		DocumentType: "charter",
		Status:       "pending",
		Title:        "Company Charter",
	})

	resp := env.postJSON(t, "/platform/reviews/organizations/"+organizationID.String()+"/legal-documents/"+documentID.String()+"/transition", map[string]any{
		"targetStatus": "rejected",
		"reviewNote":   "Missing registration stamp",
	}, token)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		ID                uuid.UUID  `json:"id"`
		OrganizationID    uuid.UUID  `json:"organizationId"`
		Status            string     `json:"status"`
		ReviewNote        *string    `json:"reviewNote"`
		ReviewerAccountID *uuid.UUID `json:"reviewerAccountId"`
		ReviewedAt        *time.Time `json:"reviewedAt"`
	}
	decodePlatformReviewsJSON(t, resp.Body, &body)

	if body.ID != documentID {
		t.Fatalf("id = %s, want %s", body.ID, documentID)
	}
	if body.OrganizationID != organizationID {
		t.Fatalf("organizationId = %s, want %s", body.OrganizationID, organizationID)
	}
	if body.Status != "rejected" {
		t.Fatalf("status = %q, want rejected", body.Status)
	}
	if body.ReviewNote == nil || *body.ReviewNote != "Missing registration stamp" {
		t.Fatalf("reviewNote = %v, want Missing registration stamp", body.ReviewNote)
	}
	if body.ReviewerAccountID == nil || *body.ReviewerAccountID != reviewerID {
		t.Fatalf("reviewerAccountId = %v, want %s", body.ReviewerAccountID, reviewerID)
	}
	if body.ReviewedAt == nil || body.ReviewedAt.IsZero() {
		t.Fatal("reviewedAt = nil/zero, want timestamp")
	}

	var (
		storedStatus     string
		storedNote       string
		storedReviewerID uuid.UUID
		storedReviewedAt time.Time
	)
	if err := env.queryRowContext(t, `
		SELECT status, review_note, reviewer_account_id, reviewed_at
		FROM org.organization_legal_documents
		WHERE organization_id = $1 AND id = $2
	`, organizationID, documentID).Scan(&storedStatus, &storedNote, &storedReviewerID, &storedReviewedAt); err != nil {
		t.Fatalf("query organization_legal_documents: %v", err)
	}
	if storedStatus != "rejected" {
		t.Fatalf("db status = %q, want rejected", storedStatus)
	}
	if storedNote != "Missing registration stamp" {
		t.Fatalf("db review_note = %q, want Missing registration stamp", storedNote)
	}
	if storedReviewerID != reviewerID {
		t.Fatalf("db reviewer_account_id = %s, want %s", storedReviewerID, reviewerID)
	}
	if storedReviewedAt.IsZero() {
		t.Fatal("db reviewed_at = zero, want timestamp")
	}

	var (
		auditAction  string
		auditStatus  string
		auditTarget  string
		auditSummary string
	)
	if err := env.queryRowContext(t, `
		SELECT action, status, target_id, summary
		FROM iam.platform_audit_events
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(&auditAction, &auditStatus, &auditTarget, &auditSummary); err != nil {
		t.Fatalf("query platform audit: %v", err)
	}
	if auditAction != "platform.organization.legal_document.review.transition" {
		t.Fatalf("audit action = %q, want platform.organization.legal_document.review.transition", auditAction)
	}
	if auditStatus != "success" {
		t.Fatalf("audit status = %q, want success", auditStatus)
	}
	if auditTarget != documentID.String() {
		t.Fatalf("audit target_id = %q, want %s", auditTarget, documentID)
	}
	if auditSummary != "from=pending to=rejected" {
		t.Fatalf("audit summary = %q, want from=pending to=rejected", auditSummary)
	}
}

func TestTransitionLegalDocumentReviewIntegrationRejectsSupportOperatorAndWritesDeniedAudit(t *testing.T) {
	env := newPlatformReviewsIntegrationEnv(t)
	supportID, token := env.createAuthenticatedPlatformAccount(t, "legal-support@example.com", platformdomain.RoleSupportOperator)
	uploaderID, _ := env.createAuthenticatedPlatformAccount(t, "legal-owner@example.com", platformdomain.RoleReviewOperator)

	organizationID := env.insertOrganization(t, "Denied Docs Org", "denied-docs-org")
	objectID := env.insertStorageObject(t, organizationID, "integration-bucket", "organizations/legal/denied-docs-org/inn.pdf", "inn.pdf", "application/pdf", 128)
	documentID := env.insertLegalDocument(t, organizationID, objectID, uploaderID, platformReviewLegalDocumentSeed{
		DocumentType: "inn_certificate",
		Status:       "pending",
		Title:        "INN Certificate",
	})

	resp := env.postJSON(t, "/platform/reviews/organizations/"+organizationID.String()+"/legal-documents/"+documentID.String()+"/transition", map[string]any{
		"targetStatus": "approved",
	}, token)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}

	var (
		actorAccountID uuid.UUID
		action         string
		targetType     string
		status         string
		summary        string
	)
	if err := env.queryRowContext(t, `
		SELECT actor_account_id, action, target_type, status, summary
		FROM iam.platform_audit_events
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(&actorAccountID, &action, &targetType, &status, &summary); err != nil {
		t.Fatalf("query denied audit: %v", err)
	}
	if actorAccountID != supportID {
		t.Fatalf("actor_account_id = %s, want %s", actorAccountID, supportID)
	}
	if action != "platform-transition-legal-document-review" {
		t.Fatalf("action = %q, want platform-transition-legal-document-review", action)
	}
	if targetType != "operation" {
		t.Fatalf("targetType = %q, want operation", targetType)
	}
	if status != "denied" {
		t.Fatalf("status = %q, want denied", status)
	}
	if !strings.Contains(summary, "platform_admin, review_operator") {
		t.Fatalf("summary = %q, want required roles message", summary)
	}
}

type platformReviewsIntegrationEnv struct {
	server       *httptest.Server
	queryDB      *sql.DB
	adminDB      *sql.DB
	appDB        *sql.DB
	dbName       string
	accountRepo  *accpg.AccountRepo
	platformRepo *platformpg.Repo
	jwtManager   *jwt.Manager
}

type platformReviewApplicationSeed struct {
	OrganizationID        uuid.UUID
	Status                string
	CompanyName           *string
	ConfirmationEmail     *string
	RepresentedCategories *string
	MinimumOrderAmount    *string
	DeliveryGeography     *string
	SalesChannels         []string
	PriceListObjectID     *uuid.UUID
	ContactFirstName      *string
	ContactLastName       *string
	ContactJobTitle       *string
	ContactEmail          *string
	ContactPhone          *string
	ReviewNote            *string
	ReviewerAccountID     *uuid.UUID
	SubmittedAt           *time.Time
	ReviewedAt            *time.Time
}

type platformReviewLegalDocumentSeed struct {
	DocumentType string
	Status       string
	Title        string
}

type platformReviewAnalysisSeed struct {
	Status               string
	Provider             string
	Summary              *string
	ExtractedFieldsJSON  string
	DetectedDocumentType *string
	ConfidenceScore      *float64
}

func newPlatformReviewsIntegrationEnv(t *testing.T) *platformReviewsIntegrationEnv {
	t.Helper()

	testDB := postgresitest.NewTempDatabase(t, "collabsphere_platform_review_it")
	postgresitest.ApplyBundledMigrations(t, testDB.QueryDB)

	conf := postgresitest.TestConfig(testDB.ConnConfig, "app")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db := bootstrap.MustOpenGormDB(conf, logger)
	bootstrap.RegisterDBHooks(db)

	accountRepo := accpg.NewAccountRepo(db)
	repo := platformpg.NewRepo(db)
	txManager := dbtx.New(db)
	service := platformapp.New(repo, repo, repo, accountRepo, repo, repo, repo, clock.NewSystemClock(), txManager, nil, nil)
	handler := NewHandler(service)
	jwtManager := jwt.NewManager(conf.Auth.JWTSecret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	root := chi.NewRouter()
	api := bootstrap.NewAPI(root, conf)
	Register(api, handler, jwtManager)

	server := httptest.NewServer(root)
	appDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}

	env := &platformReviewsIntegrationEnv{
		server:       server,
		queryDB:      testDB.QueryDB,
		adminDB:      testDB.AdminDB,
		appDB:        appDB,
		dbName:       testDB.DBName,
		accountRepo:  accountRepo,
		platformRepo: repo,
		jwtManager:   jwtManager,
	}
	t.Cleanup(func() {
		server.Close()
		_ = appDB.Close()
	})
	return env
}

func (e *platformReviewsIntegrationEnv) createAuthenticatedPlatformAccount(t *testing.T, email string, roles ...platformdomain.Role) (uuid.UUID, string) {
	t.Helper()

	emailValue, err := accdomain.NewEmail(email)
	if err != nil {
		t.Fatalf("NewEmail: %v", err)
	}
	hash, err := accdomain.NewPasswordHash("hashed:password")
	if err != nil {
		t.Fatalf("NewPasswordHash: %v", err)
	}
	account, err := accdomain.NewAccount(accdomain.NewAccountParams{
		ID:           accdomain.NewAccountID(),
		Email:        emailValue,
		PasswordHash: hash,
		Now:          time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}
	if err := e.accountRepo.Create(context.Background(), account); err != nil {
		t.Fatalf("Create account: %v", err)
	}
	if err := e.platformRepo.ReplaceRoles(context.Background(), account.ID().UUID(), roles, nil, time.Now().UTC()); err != nil {
		t.Fatalf("ReplaceRoles: %v", err)
	}

	sessionID := uuid.New()
	token, err := e.jwtManager.GenerateAccessToken(context.Background(), authdomain.NewAccountPrincipal(account.ID().UUID(), sessionID), time.Now().UTC().Add(e.jwtManager.AccessTTL()))
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	return account.ID().UUID(), token
}

func (e *platformReviewsIntegrationEnv) insertOrganization(t *testing.T, name, slug string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	now := time.Now().UTC()
	if _, err := e.queryDB.ExecContext(context.Background(), `
		INSERT INTO org.organizations (id, name, slug, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, $4, $4)
	`, id, name, slug, now); err != nil {
		t.Fatalf("insert organization: %v", err)
	}
	return id
}

func (e *platformReviewsIntegrationEnv) insertDomain(t *testing.T, organizationID uuid.UUID, hostname, kind string, isPrimary bool, verified bool) uuid.UUID {
	t.Helper()

	id := uuid.New()
	now := time.Now().UTC()
	var verifiedAt any
	if verified {
		verifiedAt = now
	}
	if _, err := e.queryDB.ExecContext(context.Background(), `
		INSERT INTO org.organization_domains (id, organization_id, hostname, kind, is_primary, verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`, id, organizationID, hostname, kind, isPrimary, verifiedAt, now); err != nil {
		t.Fatalf("insert organization domain: %v", err)
	}
	return id
}

func (e *platformReviewsIntegrationEnv) insertCooperationApplication(t *testing.T, seed platformReviewApplicationSeed) uuid.UUID {
	t.Helper()

	id := uuid.New()
	now := time.Now().UTC()
	submittedAt := seed.SubmittedAt
	if submittedAt != nil && submittedAt.Before(now) {
		submittedAt = &now
	}
	reviewedAt := seed.ReviewedAt
	if reviewedAt != nil && reviewedAt.Before(now) {
		reviewedAt = &now
	}
	if _, err := e.queryDB.ExecContext(context.Background(), `
		INSERT INTO org.cooperation_applications (
			id,
			organization_id,
			status,
			confirmation_email,
			company_name,
			represented_categories,
			minimum_order_amount,
			delivery_geography,
			sales_channels,
			price_list_object_id,
			contact_first_name,
			contact_last_name,
			contact_job_title,
			contact_email,
			contact_phone,
			review_note,
			reviewer_account_id,
			submitted_at,
			reviewed_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $20)
	`, id, seed.OrganizationID, seed.Status, seed.ConfirmationEmail, seed.CompanyName, seed.RepresentedCategories, seed.MinimumOrderAmount, seed.DeliveryGeography, nonEmptyJSONSlice(seed.SalesChannels), seed.PriceListObjectID, seed.ContactFirstName, seed.ContactLastName, seed.ContactJobTitle, seed.ContactEmail, seed.ContactPhone, seed.ReviewNote, seed.ReviewerAccountID, submittedAt, reviewedAt, now); err != nil {
		t.Fatalf("insert cooperation application: %v", err)
	}
	return id
}

func (e *platformReviewsIntegrationEnv) insertStorageObject(t *testing.T, organizationID uuid.UUID, bucket, objectKey, fileName, contentType string, sizeBytes int64) uuid.UUID {
	t.Helper()

	id := uuid.New()
	now := time.Now().UTC()
	if _, err := e.queryDB.ExecContext(context.Background(), `
		INSERT INTO storage.objects (id, organization_id, bucket, object_key, file_name, content_type, size_bytes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, organizationID, bucket, objectKey, fileName, contentType, sizeBytes, now); err != nil {
		t.Fatalf("insert storage object: %v", err)
	}
	return id
}

func (e *platformReviewsIntegrationEnv) insertLegalDocument(t *testing.T, organizationID, objectID, uploadedByAccountID uuid.UUID, seed platformReviewLegalDocumentSeed) uuid.UUID {
	t.Helper()

	id := uuid.New()
	now := time.Now().UTC()
	if _, err := e.queryDB.ExecContext(context.Background(), `
		INSERT INTO org.organization_legal_documents (
			id,
			organization_id,
			document_type,
			status,
			object_id,
			title,
			uploaded_by_account_id,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, organizationID, seed.DocumentType, seed.Status, objectID, seed.Title, uploadedByAccountID, now); err != nil {
		t.Fatalf("insert legal document: %v", err)
	}
	return id
}

func (e *platformReviewsIntegrationEnv) insertLegalDocumentAnalysis(t *testing.T, organizationID, documentID uuid.UUID, seed platformReviewAnalysisSeed) uuid.UUID {
	t.Helper()

	id := uuid.New()
	now := time.Now().UTC()
	if _, err := e.queryDB.ExecContext(context.Background(), `
		INSERT INTO org.organization_legal_document_analysis (
			id,
			document_id,
			organization_id,
			status,
			provider,
			summary,
			extracted_fields_json,
			detected_document_type,
			confidence_score,
			requested_at,
			started_at,
			completed_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10, $10, $10, $10)
	`, id, documentID, organizationID, seed.Status, seed.Provider, seed.Summary, nonEmptyJSON(seed.ExtractedFieldsJSON), seed.DetectedDocumentType, seed.ConfidenceScore, now); err != nil {
		t.Fatalf("insert legal document analysis: %v", err)
	}
	return id
}

func (e *platformReviewsIntegrationEnv) get(t *testing.T, path string, bearer string) *nethttp.Response {
	t.Helper()

	req, err := nethttp.NewRequest(nethttp.MethodGet, e.server.URL+path, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if strings.TrimSpace(bearer) != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do request: %v", err)
	}
	return resp
}

func (e *platformReviewsIntegrationEnv) postJSON(t *testing.T, path string, payload map[string]any, bearer string) *nethttp.Response {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	req, err := nethttp.NewRequest(nethttp.MethodPost, e.server.URL+path, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(bearer) != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do request: %v", err)
	}
	return resp
}

func (e *platformReviewsIntegrationEnv) queryRowContext(t *testing.T, query string, args ...any) *sql.Row {
	t.Helper()
	return e.queryDB.QueryRowContext(context.Background(), query, args...)
}

func decodePlatformReviewsJSON(t *testing.T, r io.Reader, target any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(target); err != nil {
		t.Fatalf("json.Decode: %v", err)
	}
}

func floatPtr(value float64) *float64 {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func nonEmptyJSON(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "{}"
	}
	return value
}

func nonEmptyJSONSlice(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	body, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(body)
}
