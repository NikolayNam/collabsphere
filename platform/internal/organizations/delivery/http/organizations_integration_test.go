//go:build integration

package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	nethttp "net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	accpg "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	catalogpg "github.com/NikolayNam/collabsphere/internal/catalog/repository/postgres"
	memberdomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	membershipsApp "github.com/NikolayNam/collabsphere/internal/memberships/application"
	memberpg "github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres"
	orgapp "github.com/NikolayNam/collabsphere/internal/organizations/application"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	orgpg "github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/bootstrap"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	"github.com/NikolayNam/collabsphere/internal/testutil/postgresitest"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestCreateOrganizationIntegrationRequiresAuthentication(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)

	resp := env.postJSON(t, "/v1/organizations", map[string]any{
		"name": "Acme Foods",
		"slug": "acme-foods",
	}, "")
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestCreateOrganizationIntegrationSuccess(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	accountID, token := env.createAuthenticatedAccount(t, "owner@example.com")

	resp := env.postJSON(t, "/v1/organizations", map[string]any{
		"name": "Acme Foods",
		"slug": "acme-foods",
		"domains": []map[string]any{{
			"hostname":  "acme.collabsphere.ru",
			"kind":      "subdomain",
			"isPrimary": true,
		}},
	}, token)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}

	var body struct {
		ID       uuid.UUID `json:"id"`
		Name     string    `json:"name"`
		Slug     string    `json:"slug"`
		IsActive bool      `json:"isActive"`
		Domains  []struct {
			Hostname   string `json:"hostname"`
			Kind       string `json:"kind"`
			IsPrimary  bool   `json:"isPrimary"`
			IsVerified bool   `json:"isVerified"`
		} `json:"domains"`
	}
	decodeJSON(t, resp.Body, &body)

	if body.Name != "Acme Foods" {
		t.Fatalf("name = %q, want Acme Foods", body.Name)
	}
	if body.Slug != "acme-foods" {
		t.Fatalf("slug = %q, want acme-foods", body.Slug)
	}
	if !body.IsActive {
		t.Fatal("isActive = false, want true")
	}
	if len(body.Domains) != 1 {
		t.Fatalf("domains len = %d, want 1", len(body.Domains))
	}
	if body.Domains[0].Hostname != "acme.collabsphere.ru" {
		t.Fatalf("hostname = %q, want acme.collabsphere.ru", body.Domains[0].Hostname)
	}
	if !body.Domains[0].IsVerified {
		t.Fatal("subdomain must be auto-verified on create")
	}

	var storedName, storedSlug string
	if err := env.queryRowContext(t, `SELECT name, slug FROM org.organizations WHERE id = $1`, body.ID).Scan(&storedName, &storedSlug); err != nil {
		t.Fatalf("query organization: %v", err)
	}
	if storedName != "Acme Foods" || storedSlug != "acme-foods" {
		t.Fatalf("stored organization = (%q, %q), want (Acme Foods, acme-foods)", storedName, storedSlug)
	}

	var verifiedAt time.Time
	if err := env.queryRowContext(t, `SELECT verified_at FROM org.organization_domains WHERE organization_id = $1 AND hostname = $2`, body.ID, "acme.collabsphere.ru").Scan(&verifiedAt); err != nil {
		t.Fatalf("query organization domain: %v", err)
	}
	if verifiedAt.IsZero() {
		t.Fatal("verified_at = zero, want auto-verified subdomain")
	}

	var (
		membershipRole   string
		membershipActive bool
	)
	if err := env.queryRowContext(t, `SELECT role, is_active FROM iam.memberships WHERE organization_id = $1 AND account_id = $2`, body.ID, accountID).Scan(&membershipRole, &membershipActive); err != nil {
		t.Fatalf("query membership: %v", err)
	}
	if membershipRole != "owner" {
		t.Fatalf("membership role = %q, want owner", membershipRole)
	}
	if !membershipActive {
		t.Fatal("membership is_active = false, want true")
	}
}

func TestGetOrganizationByIdIntegrationReturnsOrganization(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	_, token := env.createAuthenticatedAccount(t, "reader@example.com")

	created := env.postJSON(t, "/v1/organizations", map[string]any{
		"name": "Lookup Org",
		"slug": "lookup-org",
		"domains": []map[string]any{{
			"hostname": "lookup.collabsphere.ru",
			"kind":     "subdomain",
		}},
	}, token)
	defer created.Body.Close()
	if created.StatusCode != nethttp.StatusCreated {
		t.Fatalf("create status = %d, want 201", created.StatusCode)
	}

	var createdBody struct {
		ID uuid.UUID `json:"id"`
	}
	decodeJSON(t, created.Body, &createdBody)

	resp := env.get(t, "/v1/organizations/"+createdBody.ID.String(), "")
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		ID       uuid.UUID `json:"id"`
		Name     string    `json:"name"`
		Slug     string    `json:"slug"`
		IsActive bool      `json:"isActive"`
		Domains  []struct {
			Hostname   string `json:"hostname"`
			IsVerified bool   `json:"isVerified"`
		} `json:"domains"`
	}
	decodeJSON(t, resp.Body, &body)

	if body.ID != createdBody.ID {
		t.Fatalf("id = %s, want %s", body.ID, createdBody.ID)
	}
	if body.Name != "Lookup Org" {
		t.Fatalf("name = %q, want Lookup Org", body.Name)
	}
	if body.Slug != "lookup-org" {
		t.Fatalf("slug = %q, want lookup-org", body.Slug)
	}
	if !body.IsActive {
		t.Fatal("isActive = false, want true")
	}
	if len(body.Domains) != 1 || body.Domains[0].Hostname != "lookup.collabsphere.ru" {
		t.Fatalf("domains = %#v, want lookup.collabsphere.ru", body.Domains)
	}
	if !body.Domains[0].IsVerified {
		t.Fatal("lookup domain isVerified = false, want true")
	}
}

func TestCreateOrganizationIntegrationDuplicateSlug(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	_, token := env.createAuthenticatedAccount(t, "owner2@example.com")

	first := env.postJSON(t, "/v1/organizations", map[string]any{
		"name": "Acme Foods",
		"slug": "duplicate-org",
	}, token)
	first.Body.Close()
	if first.StatusCode != nethttp.StatusCreated {
		t.Fatalf("first status = %d, want 201", first.StatusCode)
	}

	second := env.postJSON(t, "/v1/organizations", map[string]any{
		"name": "Another Org",
		"slug": "duplicate-org",
	}, token)
	defer second.Body.Close()
	if second.StatusCode != nethttp.StatusConflict {
		t.Fatalf("second status = %d, want 409", second.StatusCode)
	}
}

func TestResolveOrganizationByHostIntegrationReturnsOrganization(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	_, token := env.createAuthenticatedAccount(t, "resolve@example.com")

	created := env.postJSON(t, "/v1/organizations", map[string]any{
		"name": "Resolve Org",
		"slug": "resolve-org",
		"domains": []map[string]any{{
			"hostname": "resolve.collabsphere.ru",
			"kind":     "subdomain",
		}},
	}, token)
	defer created.Body.Close()
	if created.StatusCode != nethttp.StatusCreated {
		t.Fatalf("create status = %d, want 201", created.StatusCode)
	}

	resp := env.get(t, "/v1/organizations/resolve-by-host?host=https://resolve.collabsphere.ru/", "")
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Name    string `json:"name"`
		Slug    string `json:"slug"`
		Domains []struct {
			Hostname string `json:"hostname"`
		} `json:"domains"`
	}
	decodeJSON(t, resp.Body, &body)

	if body.Name != "Resolve Org" {
		t.Fatalf("name = %q, want Resolve Org", body.Name)
	}
	if body.Slug != "resolve-org" {
		t.Fatalf("slug = %q, want resolve-org", body.Slug)
	}
	if len(body.Domains) != 1 || body.Domains[0].Hostname != "resolve.collabsphere.ru" {
		t.Fatalf("domains = %#v, want resolve.collabsphere.ru", body.Domains)
	}
}

func TestUploadCooperationPriceListIntegrationRequiresAuthentication(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	_, ownerToken := env.createAuthenticatedAccount(t, "price-list-owner@example.com")
	organizationID := env.createOrganization(t, ownerToken, "Price List Org", "price-list-org")

	resp := env.postMultipartFile(t, "/v1/organizations/"+organizationID.String()+"/cooperation-application/price-list", "", "file", "price-list.csv", "text/csv", []byte("sku,name\n1,Milk\n"))
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestUploadCooperationPriceListIntegrationRejectsMissingFile(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	_, ownerToken := env.createAuthenticatedAccount(t, "price-list-missing@example.com")
	organizationID := env.createOrganization(t, ownerToken, "No File Org", "no-file-org")

	resp := env.postMultipartFile(t, "/v1/organizations/"+organizationID.String()+"/cooperation-application/price-list", ownerToken, "", "", "", nil)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestUploadCooperationPriceListIntegrationRejectsViewer(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	_, ownerToken := env.createAuthenticatedAccount(t, "price-list-admin@example.com")
	organizationID := env.createOrganization(t, ownerToken, "Restricted Org", "restricted-org")

	viewerID, viewerToken := env.createAuthenticatedAccount(t, "viewer@example.com")
	env.addMember(t, organizationID, viewerID, memberdomain.MembershipRoleViewer)

	resp := env.postMultipartFile(t, "/v1/organizations/"+organizationID.String()+"/cooperation-application/price-list", viewerToken, "file", "price-list.csv", "text/csv", []byte("sku,name\n1,Milk\n"))
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
}

func TestUploadCooperationPriceListIntegrationSuccess(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	_, ownerToken := env.createAuthenticatedAccount(t, "price-list-success@example.com")
	organizationID := env.createOrganization(t, ownerToken, "Upload Org", "upload-org")

	fileContent := []byte("sku,name\n1,Milk\n2,Bread\n")
	resp := env.postMultipartFile(t, "/v1/organizations/"+organizationID.String()+"/cooperation-application/price-list", ownerToken, "file", "price-list.csv", "text/csv", fileContent)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		OrganizationID    uuid.UUID  `json:"organizationId"`
		Status            string     `json:"status"`
		PriceListObjectID *uuid.UUID `json:"priceListObjectId"`
	}
	decodeJSON(t, resp.Body, &body)

	if body.OrganizationID != organizationID {
		t.Fatalf("organizationId = %s, want %s", body.OrganizationID, organizationID)
	}
	if body.Status != "draft" {
		t.Fatalf("status = %q, want draft", body.Status)
	}
	if body.PriceListObjectID == nil || *body.PriceListObjectID == uuid.Nil {
		t.Fatal("priceListObjectId is nil, want uploaded object id")
	}

	var (
		applicationObjectID uuid.UUID
		applicationStatus   string
	)
	if err := env.queryRowContext(t, `SELECT price_list_object_id, status FROM org.cooperation_applications WHERE organization_id = $1`, organizationID).Scan(&applicationObjectID, &applicationStatus); err != nil {
		t.Fatalf("query cooperation application: %v", err)
	}
	if applicationObjectID != *body.PriceListObjectID {
		t.Fatalf("db price_list_object_id = %s, want %s", applicationObjectID, *body.PriceListObjectID)
	}
	if applicationStatus != "draft" {
		t.Fatalf("db status = %q, want draft", applicationStatus)
	}

	var (
		bucket      string
		objectKey   string
		fileName    string
		contentType *string
		sizeBytes   int64
		storedOrgID *uuid.UUID
	)
	if err := env.queryRowContext(t, `SELECT bucket, object_key, file_name, content_type, size_bytes, organization_id FROM storage.objects WHERE id = $1`, applicationObjectID).Scan(&bucket, &objectKey, &fileName, &contentType, &sizeBytes, &storedOrgID); err != nil {
		t.Fatalf("query storage object: %v", err)
	}
	if bucket != env.bucket {
		t.Fatalf("bucket = %q, want %q", bucket, env.bucket)
	}
	if !strings.HasPrefix(objectKey, "organizations/cooperation-applications/price-lists/"+organizationID.String()+"/") {
		t.Fatalf("objectKey = %q, want organizations/cooperation-applications/price-lists/%s/...", objectKey, organizationID)
	}
	if fileName != "price-list.csv" {
		t.Fatalf("fileName = %q, want price-list.csv", fileName)
	}
	if contentType == nil || *contentType != "text/csv" {
		t.Fatalf("contentType = %v, want text/csv", contentType)
	}
	if sizeBytes != int64(len(fileContent)) {
		t.Fatalf("sizeBytes = %d, want %d", sizeBytes, len(fileContent))
	}
	if storedOrgID == nil || *storedOrgID != organizationID {
		t.Fatalf("organization_id = %v, want %s", storedOrgID, organizationID)
	}

	storedUpload, ok := env.storage.object(bucket, objectKey)
	if !ok {
		t.Fatalf("storage object %s/%s not found in fake storage", bucket, objectKey)
	}
	if storedUpload.contentType != "text/csv" {
		t.Fatalf("stored contentType = %q, want text/csv", storedUpload.contentType)
	}
	if storedUpload.size != int64(len(fileContent)) {
		t.Fatalf("stored size = %d, want %d", storedUpload.size, len(fileContent))
	}
	if !bytes.Equal(storedUpload.data, fileContent) {
		t.Fatalf("stored file content = %q, want %q", storedUpload.data, fileContent)
	}
}

func TestGetOrganizationLegalDocumentVerificationIntegrationApproved(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	accountID, token := env.createAuthenticatedAccount(t, "doc-verify-approved@example.com")
	organizationID := env.createOrganization(t, token, "Verified Docs Org", "verified-docs-org")

	objectID := env.insertStorageObject(t, organizationID, "legal/inn.pdf", "inn.pdf", "application/pdf", 128)
	documentID := env.insertLegalDocument(t, organizationID, objectID, accountID, "inn_certificate", "INN Certificate")
	env.insertLegalDocumentAnalysis(t, organizationID, documentID, "completed", "generic-http", json.RawMessage(`{"inn":"7701234567","companyName":"Acme","registrationDate":"2022-01-01"}`), stringPtr("inn_certificate"), floatPtr(0.98), nil)

	resp := env.get(t, "/v1/organizations/"+organizationID.String()+"/legal-documents/"+documentID.String()+"/verification", token)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		DocumentID     uuid.UUID `json:"documentId"`
		Verdict        string    `json:"verdict"`
		DocumentType   string    `json:"documentType"`
		RequiredFields []string  `json:"requiredFields"`
		MissingFields  []string  `json:"missingFields"`
		Issues         []struct {
			Code string `json:"code"`
		} `json:"issues"`
	}
	decodeJSON(t, resp.Body, &body)

	if body.DocumentID != documentID {
		t.Fatalf("documentId = %s, want %s", body.DocumentID, documentID)
	}
	if body.Verdict != "approved" {
		t.Fatalf("verdict = %q, want approved", body.Verdict)
	}
	if body.DocumentType != "inn_certificate" {
		t.Fatalf("documentType = %q, want inn_certificate", body.DocumentType)
	}
	if len(body.RequiredFields) != 3 {
		t.Fatalf("requiredFields len = %d, want 3", len(body.RequiredFields))
	}
	if len(body.MissingFields) != 0 {
		t.Fatalf("missingFields = %v, want empty", body.MissingFields)
	}
	if len(body.Issues) != 0 {
		t.Fatalf("issues = %+v, want empty", body.Issues)
	}
}

func TestGetOrganizationLegalDocumentVerificationIntegrationManualReviewWhenAnalysisFailed(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	accountID, token := env.createAuthenticatedAccount(t, "doc-verify-failed@example.com")
	organizationID := env.createOrganization(t, token, "Failed Docs Org", "failed-docs-org")

	objectID := env.insertStorageObject(t, organizationID, "legal/ogrn.pdf", "ogrn.pdf", "application/pdf", 128)
	documentID := env.insertLegalDocument(t, organizationID, objectID, accountID, "ogrn_extract", "OGRN Extract")
	env.insertLegalDocumentAnalysis(t, organizationID, documentID, "failed", "generic-http", json.RawMessage(`{}`), nil, nil, stringPtr("Provider timeout"))

	resp := env.get(t, "/v1/organizations/"+organizationID.String()+"/legal-documents/"+documentID.String()+"/verification", token)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Verdict string `json:"verdict"`
		Issues  []struct {
			Code string `json:"code"`
		} `json:"issues"`
	}
	decodeJSON(t, resp.Body, &body)

	if body.Verdict != "manual_review" {
		t.Fatalf("verdict = %q, want manual_review", body.Verdict)
	}
	if len(body.Issues) != 1 || body.Issues[0].Code != "analysis_failed" {
		t.Fatalf("issues = %+v, want analysis_failed", body.Issues)
	}
}

func TestGetOrganizationLegalDocumentVerificationIntegrationRejectedOnTypeMismatch(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	accountID, token := env.createAuthenticatedAccount(t, "doc-verify-mismatch@example.com")
	organizationID := env.createOrganization(t, token, "Mismatch Docs Org", "mismatch-docs-org")

	objectID := env.insertStorageObject(t, organizationID, "legal/inn.pdf", "inn.pdf", "application/pdf", 128)
	documentID := env.insertLegalDocument(t, organizationID, objectID, accountID, "inn_certificate", "INN Certificate")
	env.insertLegalDocumentAnalysis(t, organizationID, documentID, "completed", "generic-http", json.RawMessage(`{"companyName":"Acme","registrationDate":"2022-01-01","inn":"7701234567"}`), stringPtr("charter"), floatPtr(0.99), nil)

	resp := env.get(t, "/v1/organizations/"+organizationID.String()+"/legal-documents/"+documentID.String()+"/verification", token)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Verdict string `json:"verdict"`
		Issues  []struct {
			Code string `json:"code"`
		} `json:"issues"`
	}
	decodeJSON(t, resp.Body, &body)

	if body.Verdict != "rejected" {
		t.Fatalf("verdict = %q, want rejected", body.Verdict)
	}
	if len(body.Issues) == 0 || body.Issues[0].Code != "document_type_mismatch" {
		t.Fatalf("issues = %+v, want document_type_mismatch", body.Issues)
	}
}

func TestGetOrganizationKYCRequirementsIntegrationCurrentlyDue(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	_, token := env.createAuthenticatedAccount(t, "kyc-due@example.com")
	organizationID := env.createOrganization(t, token, "KYC Due Org", "kyc-due-org")

	resp := env.get(t, "/v1/organizations/"+organizationID.String()+"/kyc/requirements", token)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		OrganizationID uuid.UUID `json:"organizationId"`
		Status         string    `json:"status"`
		CurrentlyDue   []struct {
			Code string `json:"code"`
		} `json:"currentlyDue"`
	}
	decodeJSON(t, resp.Body, &body)

	if body.OrganizationID != organizationID {
		t.Fatalf("organizationId = %s, want %s", body.OrganizationID, organizationID)
	}
	if body.Status != "currently_due" {
		t.Fatalf("status = %q, want currently_due", body.Status)
	}
	if len(body.CurrentlyDue) != 2 {
		t.Fatalf("currentlyDue len = %d, want 2", len(body.CurrentlyDue))
	}
}

func TestGetOrganizationKYCRequirementsIntegrationPendingVerification(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	accountID, token := env.createAuthenticatedAccount(t, "kyc-pending@example.com")
	organizationID := env.createOrganization(t, token, "KYC Pending Org", "kyc-pending-org")

	env.insertCooperationApplication(t, organizationID, "submitted", accountID)
	objectID := env.insertStorageObject(t, organizationID, "legal/charter.pdf", "charter.pdf", "application/pdf", 128)
	documentID := env.insertLegalDocument(t, organizationID, objectID, accountID, "charter", "Company Charter")
	env.insertLegalDocumentAnalysis(t, organizationID, documentID, "completed", "generic-http", json.RawMessage(`{"companyName":"Acme"}`), stringPtr("charter"), floatPtr(0.96), nil)

	resp := env.get(t, "/v1/organizations/"+organizationID.String()+"/kyc/requirements", token)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Status              string `json:"status"`
		PendingVerification []struct {
			Code string `json:"code"`
		} `json:"pendingVerification"`
	}
	decodeJSON(t, resp.Body, &body)

	if body.Status != "pending_verification" {
		t.Fatalf("status = %q, want pending_verification", body.Status)
	}
	if len(body.PendingVerification) != 2 {
		t.Fatalf("pendingVerification len = %d, want 2", len(body.PendingVerification))
	}
}

func TestGetOrganizationKYCRequirementsIntegrationNeedsInfo(t *testing.T) {
	env := newOrganizationsIntegrationEnv(t)
	accountID, token := env.createAuthenticatedAccount(t, "kyc-needs-info@example.com")
	organizationID := env.createOrganization(t, token, "KYC Needs Info Org", "kyc-needs-info-org")

	env.insertCooperationApplication(t, organizationID, "approved", accountID)
	objectID := env.insertStorageObject(t, organizationID, "legal/inn.pdf", "inn.pdf", "application/pdf", 128)
	documentID := env.insertLegalDocument(t, organizationID, objectID, accountID, "inn_certificate", "INN Certificate")
	if _, err := env.queryDB.ExecContext(context.Background(), `
		UPDATE org.organization_legal_documents
		SET status = 'rejected', review_note = 'Upload clearer scan', reviewer_account_id = $2, reviewed_at = NOW()
		WHERE id = $1
	`, documentID, accountID); err != nil {
		t.Fatalf("reject legal document: %v", err)
	}

	resp := env.get(t, "/v1/organizations/"+organizationID.String()+"/kyc/requirements", token)
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Status         string  `json:"status"`
		DisabledReason *string `json:"disabledReason"`
		Errors         []struct {
			Code   string  `json:"code"`
			Reason *string `json:"reason"`
		} `json:"errors"`
	}
	decodeJSON(t, resp.Body, &body)

	if body.Status != "needs_info" {
		t.Fatalf("status = %q, want needs_info", body.Status)
	}
	if body.DisabledReason == nil || *body.DisabledReason != "requirements_past_due" {
		t.Fatalf("disabledReason = %v, want requirements_past_due", body.DisabledReason)
	}
	if len(body.Errors) != 1 || body.Errors[0].Code != "legal_document.review" {
		t.Fatalf("errors = %+v, want legal_document.review", body.Errors)
	}
	if body.Errors[0].Reason == nil || *body.Errors[0].Reason != "Upload clearer scan" {
		t.Fatalf("error reason = %v, want Upload clearer scan", body.Errors[0].Reason)
	}
}

type organizationsIntegrationEnv struct {
	server      *httptest.Server
	queryDB     *sql.DB
	adminDB     *sql.DB
	appDB       *sql.DB
	dbName      string
	accountRepo *accpg.AccountRepo
	memberRepo  *memberpg.MembershipRepo
	jwtManager  *jwt.Manager
	storage     *fakeOrganizationStorage
	bucket      string
}

func newOrganizationsIntegrationEnv(t *testing.T) *organizationsIntegrationEnv {
	t.Helper()

	testDB := postgresitest.NewTempDatabase(t, "collabsphere_org_it")
	postgresitest.ApplyBundledMigrations(t, testDB.QueryDB)

	conf := postgresitest.TestConfig(testDB.ConnConfig, "app")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db := bootstrap.MustOpenGormDB(conf, logger)
	bootstrap.RegisterDBHooks(db)

	accountRepo := accpg.NewAccountRepo(db)
	orgRepo := orgpg.NewOrganizationRepo(db)
	membershipRepo := memberpg.NewMembershipRepo(db)
	roleRepo := memberpg.NewOrganizationRoleRepo(db)
	roleResolver := membershipsApp.NewRoleResolverAdapter(roleRepo)
	storage := newFakeOrganizationStorage()
	const bucket = "integration-bucket"
	txManager := dbtx.New(db)
	catalogRepo := catalogpg.NewCatalogRepo(db)
	service := orgapp.New(orgRepo, membershipRepo, roleResolver, nil, catalogRepo, txManager, clock.NewSystemClock(), storage, bucket, nil, "", nil)
	handler := NewHandler(service)
	jwtManager := jwt.NewManager(conf.Auth.JWTSecret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)

	root := chi.NewRouter()
	apiV1 := chi.NewRouter()
	root.Mount("/v1", apiV1)
	api := bootstrap.NewAPI(apiV1, conf)
	Register(api, handler, jwtManager)

	server := httptest.NewServer(root)
	appDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}

	env := &organizationsIntegrationEnv{
		server:      server,
		queryDB:     testDB.QueryDB,
		adminDB:     testDB.AdminDB,
		appDB:       appDB,
		dbName:      testDB.DBName,
		accountRepo: accountRepo,
		memberRepo:  membershipRepo,
		jwtManager:  jwtManager,
		storage:     storage,
		bucket:      bucket,
	}
	t.Cleanup(func() {
		server.Close()
		_ = appDB.Close()
	})
	return env
}

func (e *organizationsIntegrationEnv) createAuthenticatedAccount(t *testing.T, email string) (uuid.UUID, string) {
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

	sessionID := uuid.New()
	token, err := e.jwtManager.GenerateAccessToken(context.Background(), authdomain.NewAccountPrincipal(account.ID().UUID(), sessionID), time.Now().UTC().Add(e.jwtManager.AccessTTL()))
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	return account.ID().UUID(), token
}

func (e *organizationsIntegrationEnv) createOrganization(t *testing.T, bearer, name, slug string) uuid.UUID {
	t.Helper()

	resp := e.postJSON(t, "/v1/organizations", map[string]any{
		"name": name,
		"slug": slug,
	}, bearer)
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusCreated {
		t.Fatalf("create organization status = %d, want 201", resp.StatusCode)
	}

	var body struct {
		ID uuid.UUID `json:"id"`
	}
	decodeJSON(t, resp.Body, &body)
	return body.ID
}

func (e *organizationsIntegrationEnv) addMember(t *testing.T, organizationID, accountID uuid.UUID, role memberdomain.MembershipRole) {
	t.Helper()

	orgID, err := orgdomain.OrganizationIDFromUUID(organizationID)
	if err != nil {
		t.Fatalf("OrganizationIDFromUUID: %v", err)
	}
	accID, err := accdomain.AccountIDFromUUID(accountID)
	if err != nil {
		t.Fatalf("AccountIDFromUUID: %v", err)
	}
	membership, err := memberdomain.NewMembership(memberdomain.NewMembershipParams{
		OrganizationID: orgID,
		AccountID:      accID,
		Role:           role,
		Now:            time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("NewMembership: %v", err)
	}
	if err := e.memberRepo.AddMember(context.Background(), orgID, membership); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
}

func (e *organizationsIntegrationEnv) insertStorageObject(t *testing.T, organizationID uuid.UUID, objectKey, fileName, contentType string, sizeBytes int64) uuid.UUID {
	t.Helper()

	id := uuid.New()
	now := time.Now().UTC()
	if _, err := e.queryDB.ExecContext(context.Background(), `
		INSERT INTO storage.objects (id, organization_id, bucket, object_key, file_name, content_type, size_bytes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, organizationID, e.bucket, objectKey, fileName, contentType, sizeBytes, now); err != nil {
		t.Fatalf("insert storage object: %v", err)
	}
	return id
}

func (e *organizationsIntegrationEnv) insertLegalDocument(t *testing.T, organizationID, objectID, uploadedByAccountID uuid.UUID, documentType, title string) uuid.UUID {
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
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, 'pending', $4, $5, $6, $7, $7)
	`, id, organizationID, documentType, objectID, title, uploadedByAccountID, now); err != nil {
		t.Fatalf("insert legal document: %v", err)
	}
	return id
}

func (e *organizationsIntegrationEnv) insertCooperationApplication(t *testing.T, organizationID uuid.UUID, status string, reviewerAccountID uuid.UUID) uuid.UUID {
	t.Helper()

	id := uuid.New()
	now := time.Now().UTC()
	priceListObjectID := e.insertStorageObject(t, organizationID, "cooperation/price-list.csv", "price-list.csv", "text/csv", 64)
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
			contact_first_name,
			contact_last_name,
			contact_job_title,
			price_list_object_id,
			contact_email,
			contact_phone,
			reviewer_account_id,
			submitted_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, 'confirm@example.com', 'Acme', 'Food', '1000', 'Moscow', '["Retail"]'::jsonb, 'Ivan', 'Petrov', 'Manager', $4, 'sales@example.com', '+79990000000', $5, $6, $6, $6)
	`, id, organizationID, status, priceListObjectID, reviewerAccountID, now); err != nil {
		t.Fatalf("insert cooperation application: %v", err)
	}
	return id
}

func (e *organizationsIntegrationEnv) insertLegalDocumentAnalysis(t *testing.T, organizationID, documentID uuid.UUID, status, provider string, fields json.RawMessage, detectedType *string, confidence *float64, lastError *string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	now := time.Now().UTC()
	if len(fields) == 0 {
		fields = json.RawMessage(`{}`)
	}
	if _, err := e.queryDB.ExecContext(context.Background(), `
		INSERT INTO org.organization_legal_document_analysis (
			id,
			document_id,
			organization_id,
			status,
			provider,
			extracted_fields_json,
			detected_document_type,
			confidence_score,
			requested_at,
			started_at,
			completed_at,
			updated_at,
			last_error
		)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $9, $10, $9, $11)
	`, id, documentID, organizationID, status, provider, string(fields), detectedType, confidence, now, completedAtForAnalysisStatus(status, now), lastError); err != nil {
		t.Fatalf("insert legal document analysis: %v", err)
	}
	return id
}

func (e *organizationsIntegrationEnv) postJSON(t *testing.T, path string, payload map[string]any, bearer string) *nethttp.Response {
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

func (e *organizationsIntegrationEnv) postMultipartFile(t *testing.T, path, bearer, fieldName, fileName, contentType string, data []byte) *nethttp.Response {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if strings.TrimSpace(fieldName) != "" {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileName))
		if strings.TrimSpace(contentType) != "" {
			header.Set("Content-Type", contentType)
		}
		part, err := writer.CreatePart(header)
		if err != nil {
			t.Fatalf("CreatePart: %v", err)
		}
		if len(data) > 0 {
			if _, err := part.Write(data); err != nil {
				t.Fatalf("Write multipart file: %v", err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart close: %v", err)
	}

	req, err := nethttp.NewRequest(nethttp.MethodPost, e.server.URL+path, &body)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if strings.TrimSpace(bearer) != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do request: %v", err)
	}
	return resp
}

func (e *organizationsIntegrationEnv) get(t *testing.T, path string, bearer string) *nethttp.Response {
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

func (e *organizationsIntegrationEnv) queryRowContext(t *testing.T, query string, args ...any) *sql.Row {
	t.Helper()
	return e.queryDB.QueryRowContext(context.Background(), query, args...)
}

func decodeJSON(t *testing.T, r io.Reader, target any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(target); err != nil {
		t.Fatalf("json.Decode: %v", err)
	}
}

func completedAtForAnalysisStatus(status string, now time.Time) any {
	switch strings.TrimSpace(status) {
	case "completed":
		return now
	default:
		return nil
	}
}

func floatPtr(value float64) *float64 {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

type fakeStoredObject struct {
	contentType string
	size        int64
	data        []byte
}

type fakeOrganizationStorage struct {
	mu      sync.Mutex
	objects map[string]fakeStoredObject
}

func newFakeOrganizationStorage() *fakeOrganizationStorage {
	return &fakeOrganizationStorage{
		objects: make(map[string]fakeStoredObject),
	}
}

func (s *fakeOrganizationStorage) PresignPutObject(_ context.Context, bucket, objectKey string) (string, time.Time, error) {
	return "https://storage.test/" + bucket + "/" + objectKey, time.Now().UTC().Add(15 * time.Minute), nil
}

func (s *fakeOrganizationStorage) PutObject(_ context.Context, bucket, objectKey string, body io.Reader, size int64, contentType string) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[s.objectKey(bucket, objectKey)] = fakeStoredObject{
		contentType: contentType,
		size:        size,
		data:        append([]byte(nil), data...),
	}
	return nil
}

func (s *fakeOrganizationStorage) ReadObject(_ context.Context, bucket, objectKey string) (io.ReadCloser, error) {
	object, ok := s.object(bucket, objectKey)
	if !ok {
		return nil, os.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader(object.data)), nil
}

func (s *fakeOrganizationStorage) object(bucket, objectKey string) (fakeStoredObject, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	object, ok := s.objects[s.objectKey(bucket, objectKey)]
	return object, ok
}

func (s *fakeOrganizationStorage) objectKey(bucket, objectKey string) string {
	return bucket + "|" + objectKey
}
