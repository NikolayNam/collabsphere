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
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	accountsapp "github.com/NikolayNam/collabsphere/internal/accounts/application"
	accountshttp "github.com/NikolayNam/collabsphere/internal/accounts/delivery/http"
	accpg "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	catalogpg "github.com/NikolayNam/collabsphere/internal/catalog/repository/postgres"
	membershipsapp "github.com/NikolayNam/collabsphere/internal/memberships/application"
	memberdomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	memberspg "github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres"
	orgapp "github.com/NikolayNam/collabsphere/internal/organizations/application"
	orghttp "github.com/NikolayNam/collabsphere/internal/organizations/delivery/http"
	orgpg "github.com/NikolayNam/collabsphere/internal/organizations/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/bootstrap"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/bcrypt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/tokens"
	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	"github.com/NikolayNam/collabsphere/internal/testutil/postgresitest"
	uploadpg "github.com/NikolayNam/collabsphere/internal/uploads/repository/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestMembershipsIntegrationAddAndListMembers(t *testing.T) {
	env := newMembershipsIntegrationEnv(t)

	owner := env.createAccount(t, "owner@example.com", "Secret123")
	member := env.createAccount(t, "member@example.com", "Secret123")
	ownerBearer := env.issueAccessToken(t, owner.ID)
	organizationID := env.createOrganization(t, ownerBearer, "Acme Foods", "acme-foods")

	addResp := env.postJSON(t, fmt.Sprintf("/v1/organizations/%s/members", organizationID), map[string]any{
		"accountId": member.ID.String(),
		"role":      "admin",
	}, ownerBearer)
	defer addResp.Body.Close()
	if addResp.StatusCode != nethttp.StatusCreated {
		t.Fatalf("add member status = %d, want 201", addResp.StatusCode)
	}

	var addBody membershipPayload
	decodeJSONMemberships(t, addResp.Body, &addBody)
	if addBody.AccountID != member.ID {
		t.Fatalf("accountId = %s, want %s", addBody.AccountID, member.ID)
	}
	if addBody.Role != "admin" {
		t.Fatalf("role = %q, want admin", addBody.Role)
	}

	listResp := env.get(t, fmt.Sprintf("/v1/organizations/%s/members", organizationID), ownerBearer)
	defer listResp.Body.Close()
	if listResp.StatusCode != nethttp.StatusOK {
		t.Fatalf("list members status = %d, want 200", listResp.StatusCode)
	}

	var listBody struct {
		Members []membershipPayload `json:"members"`
	}
	decodeJSONMemberships(t, listResp.Body, &listBody)
	if len(listBody.Members) != 2 {
		t.Fatalf("members len = %d, want 2", len(listBody.Members))
	}
}

func TestMembershipsIntegrationViewerCannotManageMembers(t *testing.T) {
	env := newMembershipsIntegrationEnv(t)

	owner := env.createAccount(t, "owner2@example.com", "Secret123")
	viewer := env.createAccount(t, "viewer@example.com", "Secret123")
	target := env.createAccount(t, "target@example.com", "Secret123")
	ownerBearer := env.issueAccessToken(t, owner.ID)
	organizationID := env.createOrganization(t, ownerBearer, "Viewer Org", "viewer-org")

	resp := env.postJSON(t, fmt.Sprintf("/v1/organizations/%s/members", organizationID), map[string]any{
		"accountId": viewer.ID.String(),
		"role":      "viewer",
	}, ownerBearer)
	resp.Body.Close()
	if resp.StatusCode != nethttp.StatusCreated {
		t.Fatalf("seed viewer status = %d, want 201", resp.StatusCode)
	}

	viewerBearer := env.issueAccessToken(t, viewer.ID)
	forbiddenResp := env.postJSON(t, fmt.Sprintf("/v1/organizations/%s/members", organizationID), map[string]any{
		"accountId": target.ID.String(),
		"role":      "member",
	}, viewerBearer)
	defer forbiddenResp.Body.Close()
	if forbiddenResp.StatusCode != nethttp.StatusForbidden {
		t.Fatalf("viewer add member status = %d, want 403", forbiddenResp.StatusCode)
	}
}

func TestMembershipsIntegrationRejectsRemovingLastActiveOwner(t *testing.T) {
	env := newMembershipsIntegrationEnv(t)

	owner := env.createAccount(t, "owner3@example.com", "Secret123")
	ownerBearer := env.issueAccessToken(t, owner.ID)
	organizationID := env.createOrganization(t, ownerBearer, "Owner Org", "owner-org")

	membersResp := env.get(t, fmt.Sprintf("/v1/organizations/%s/members", organizationID), ownerBearer)
	defer membersResp.Body.Close()
	if membersResp.StatusCode != nethttp.StatusOK {
		t.Fatalf("list members status = %d, want 200", membersResp.StatusCode)
	}

	var listBody struct {
		Members []membershipPayload `json:"members"`
	}
	decodeJSONMemberships(t, membersResp.Body, &listBody)
	if len(listBody.Members) != 1 {
		t.Fatalf("members len = %d, want 1", len(listBody.Members))
	}

	removeResp := env.delete(t, fmt.Sprintf("/v1/organizations/%s/members/%s", organizationID, listBody.Members[0].ID), ownerBearer)
	defer removeResp.Body.Close()
	if removeResp.StatusCode != nethttp.StatusConflict {
		t.Fatalf("remove owner status = %d, want 409", removeResp.StatusCode)
	}
}

func TestMembershipsIntegrationCreateListAndAcceptInvitation(t *testing.T) {
	env := newMembershipsIntegrationEnv(t)

	owner := env.createAccount(t, "owner-invite@example.com", "Secret123")
	invitee := env.createAccount(t, "invitee@example.com", "Secret123")
	ownerBearer := env.issueAccessToken(t, owner.ID)
	inviteeBearer := env.issueAccessToken(t, invitee.ID)
	organizationID := env.createOrganization(t, ownerBearer, "Invite Org", "invite-org")

	createResp := env.postJSON(t, fmt.Sprintf("/v1/organizations/%s/invitations", organizationID), map[string]any{
		"email": "invitee@example.com",
		"role":  "member",
	}, ownerBearer)
	defer createResp.Body.Close()
	if createResp.StatusCode != nethttp.StatusCreated {
		t.Fatalf("create invitation status = %d, want 201", createResp.StatusCode)
	}

	var createdInvitation invitationPayload
	decodeJSONMemberships(t, createResp.Body, &createdInvitation)
	if createdInvitation.Token == nil || strings.TrimSpace(*createdInvitation.Token) == "" {
		t.Fatal("invitation token is empty")
	}
	if createdInvitation.Status != "pending" {
		t.Fatalf("invitation status = %q, want pending", createdInvitation.Status)
	}

	listResp := env.get(t, fmt.Sprintf("/v1/organizations/%s/invitations", organizationID), ownerBearer)
	defer listResp.Body.Close()
	if listResp.StatusCode != nethttp.StatusOK {
		t.Fatalf("list invitations status = %d, want 200", listResp.StatusCode)
	}

	var listBody struct {
		Invitations []invitationPayload `json:"invitations"`
	}
	decodeJSONMemberships(t, listResp.Body, &listBody)
	if len(listBody.Invitations) != 1 {
		t.Fatalf("invitations len = %d, want 1", len(listBody.Invitations))
	}

	acceptResp := env.post(t, fmt.Sprintf("/v1/invitations/%s/accept", *createdInvitation.Token), inviteeBearer)
	defer acceptResp.Body.Close()
	if acceptResp.StatusCode != nethttp.StatusOK {
		t.Fatalf("accept invitation status = %d, want 200", acceptResp.StatusCode)
	}

	var acceptBody struct {
		Invitation invitationPayload `json:"invitation"`
		Member     membershipPayload `json:"member"`
	}
	decodeJSONMemberships(t, acceptResp.Body, &acceptBody)
	if acceptBody.Invitation.Status != "accepted" {
		t.Fatalf("accepted invitation status = %q, want accepted", acceptBody.Invitation.Status)
	}
	if acceptBody.Member.AccountID != invitee.ID {
		t.Fatalf("accepted member accountId = %s, want %s", acceptBody.Member.AccountID, invitee.ID)
	}
	if acceptBody.Member.Role != "member" {
		t.Fatalf("accepted member role = %q, want member", acceptBody.Member.Role)
	}

	var auditCount int
	if err := env.queryRowContext(t, `
		SELECT COUNT(*)
		FROM iam.organization_access_audit_events
		WHERE organization_id = $1
	`, organizationID).Scan(&auditCount); err != nil {
		t.Fatalf("count access audit events: %v", err)
	}
	if auditCount < 3 {
		t.Fatalf("audit count = %d, want at least 3", auditCount)
	}
}

func TestMembershipsIntegrationRejectsInvitationEmailMismatch(t *testing.T) {
	env := newMembershipsIntegrationEnv(t)

	owner := env.createAccount(t, "owner-mismatch@example.com", "Secret123")
	_ = env.createAccount(t, "right-invitee@example.com", "Secret123")
	other := env.createAccount(t, "wrong-invitee@example.com", "Secret123")
	ownerBearer := env.issueAccessToken(t, owner.ID)
	otherBearer := env.issueAccessToken(t, other.ID)
	organizationID := env.createOrganization(t, ownerBearer, "Mismatch Org", "mismatch-org")

	createResp := env.postJSON(t, fmt.Sprintf("/v1/organizations/%s/invitations", organizationID), map[string]any{
		"email": "right-invitee@example.com",
		"role":  "member",
	}, ownerBearer)
	defer createResp.Body.Close()
	if createResp.StatusCode != nethttp.StatusCreated {
		t.Fatalf("create invitation status = %d, want 201", createResp.StatusCode)
	}

	var createdInvitation invitationPayload
	decodeJSONMemberships(t, createResp.Body, &createdInvitation)
	if createdInvitation.Token == nil || strings.TrimSpace(*createdInvitation.Token) == "" {
		t.Fatal("invitation token is empty")
	}

	acceptResp := env.post(t, fmt.Sprintf("/v1/invitations/%s/accept", *createdInvitation.Token), otherBearer)
	defer acceptResp.Body.Close()
	if acceptResp.StatusCode != nethttp.StatusForbidden {
		t.Fatalf("accept mismatch invitation status = %d, want 403", acceptResp.StatusCode)
	}

	var problem struct {
		Code string `json:"code"`
	}
	decodeJSONMemberships(t, acceptResp.Body, &problem)
	if problem.Code != "ORGANIZATION_INVITATION_EMAIL_MISMATCH" {
		t.Fatalf("problem code = %q, want ORGANIZATION_INVITATION_EMAIL_MISMATCH", problem.Code)
	}

}

type membershipsIntegrationEnv struct {
	server     *httptest.Server
	queryDB    *sql.DB
	adminDB    *sql.DB
	appDB      *sql.DB
	dbName     string
	jwtManager *jwt.Manager
}

type membershipAccount struct {
	ID uuid.UUID
}

type membershipPayload struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organizationId"`
	AccountID      uuid.UUID `json:"accountId"`
	Role           string    `json:"role"`
	IsActive       bool      `json:"isActive"`
}

type invitationPayload struct {
	ID                  uuid.UUID  `json:"id"`
	OrganizationID      uuid.UUID  `json:"organizationId"`
	Email               string     `json:"email"`
	Role                string     `json:"role"`
	Status              string     `json:"status"`
	Token               *string    `json:"token"`
	InviterAccountID    uuid.UUID  `json:"inviterAccountId"`
	AcceptedByAccountID *uuid.UUID `json:"acceptedByAccountId"`
	AcceptedAt          *time.Time `json:"acceptedAt"`
	ExpiresAt           time.Time  `json:"expiresAt"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           *time.Time `json:"updatedAt"`
}

func newMembershipsIntegrationEnv(t *testing.T) *membershipsIntegrationEnv {
	t.Helper()

	testDB := postgresitest.NewTempDatabase(t, "collabsphere_memberships_it")
	postgresitest.ApplyBundledMigrations(t, testDB.QueryDB)

	conf := postgresitest.TestConfig(testDB.ConnConfig, "app")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db := bootstrap.MustOpenGormDB(conf, logger)
	bootstrap.RegisterDBHooks(db)

	accountRepo := accpg.NewAccountRepo(db)
	organizationRepo := orgpg.NewOrganizationRepo(db)
	membershipRepo := memberspg.NewMembershipRepo(db)
	categoryRepo := catalogpg.NewProductCategoryRepo(db)
	uploadRepo := uploadpg.NewRepo(db)

	passwordHasher := bcrypt.NewBcryptHasher()
	clk := clock.NewSystemClock()
	jwtManager := jwt.NewManager(conf.Auth.JWTSecret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)
	tokenGenerator := tokens.NewGenerator()

	accountService := accountsapp.New(accountRepo, passwordHasher, clk, nil, "")
	accountHandler := accountshttp.NewHandler(accountService, conf.Auth.LocalSignupEnabled)

	catalogRepo := catalogpg.NewCatalogRepo(db)
	roleRepo := memberspg.NewOrganizationRoleRepo(db)
	roleResolver := membershipsapp.NewRoleResolverAdapter(roleRepo)
	organizationService := orgapp.New(organizationRepo, membershipRepo, roleResolver, categoryRepo, catalogRepo, dbtx.New(db), clk, nil, "", nil, "", uploadRepo)
	organizationHandler := orghttp.NewHandler(organizationService)

	roleRepo := memberspg.NewOrganizationRoleRepo(db)
	membershipService := membershipsapp.New(membershipRepo, roleRepo, organizationRepo, accountRepo, dbtx.New(db), tokenGenerator, memberspg.NewAccessAuditRepo(db), clk, 7*24*time.Hour)
	membershipHandler := NewHandler(membershipService)

	root := chi.NewRouter()
	apiV1 := chi.NewRouter()
	root.Mount("/v1", apiV1)
	api := bootstrap.NewAPI(apiV1, conf)
	accountshttp.Register(api, accountHandler, jwtManager)
	orghttp.Register(api, organizationHandler, jwtManager)
	Register(api, membershipHandler, jwtManager)

	server := httptest.NewServer(root)
	appDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}

	env := &membershipsIntegrationEnv{
		server:     server,
		queryDB:    testDB.QueryDB,
		adminDB:    testDB.AdminDB,
		appDB:      appDB,
		dbName:     testDB.DBName,
		jwtManager: jwtManager,
	}
	t.Cleanup(func() {
		server.Close()
		_ = appDB.Close()
	})
	return env
}

func (e *membershipsIntegrationEnv) createAccount(t *testing.T, email, password string) membershipAccount {
	t.Helper()

	resp := e.postJSON(t, "/v1/accounts", map[string]any{
		"email":    email,
		"password": password,
	}, "")
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusCreated {
		t.Fatalf("create account status = %d, want 201", resp.StatusCode)
	}

	var body struct {
		ID uuid.UUID `json:"id"`
	}
	decodeJSONMemberships(t, resp.Body, &body)
	return membershipAccount{ID: body.ID}
}

func (e *membershipsIntegrationEnv) createOrganization(t *testing.T, bearer, name, slug string) uuid.UUID {
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
	decodeJSONMemberships(t, resp.Body, &body)
	return body.ID
}

func (e *membershipsIntegrationEnv) issueAccessToken(t *testing.T, accountID uuid.UUID) string {
	t.Helper()

	token, err := e.jwtManager.GenerateAccessToken(context.Background(), authdomain.NewAccountPrincipal(accountID, uuid.New()), time.Now().Add(e.jwtManager.AccessTTL()))
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	return token
}

func (e *membershipsIntegrationEnv) postJSON(t *testing.T, path string, payload map[string]any, accessToken string) *nethttp.Response {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	req, err := nethttp.NewRequest(nethttp.MethodPost, e.server.URL+path, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(accessToken) != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	return resp
}

func (e *membershipsIntegrationEnv) get(t *testing.T, path, accessToken string) *nethttp.Response {
	t.Helper()

	req, err := nethttp.NewRequest(nethttp.MethodGet, e.server.URL+path, nil)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	if strings.TrimSpace(accessToken) != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	return resp
}

func (e *membershipsIntegrationEnv) post(t *testing.T, path, accessToken string) *nethttp.Response {
	t.Helper()

	req, err := nethttp.NewRequest(nethttp.MethodPost, e.server.URL+path, nil)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	if strings.TrimSpace(accessToken) != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	return resp
}

func (e *membershipsIntegrationEnv) delete(t *testing.T, path, accessToken string) *nethttp.Response {
	t.Helper()

	req, err := nethttp.NewRequest(nethttp.MethodDelete, e.server.URL+path, nil)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	if strings.TrimSpace(accessToken) != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.Do: %v", err)
	}
	return resp
}

func decodeJSONMemberships(t *testing.T, r io.Reader, target any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(target); err != nil {
		t.Fatalf("json.Decode: %v", err)
	}
}

func (e *membershipsIntegrationEnv) queryRowContext(t *testing.T, query string, args ...any) *sql.Row {
	t.Helper()
	return e.queryDB.QueryRowContext(context.Background(), query, args...)
}

var _ = memberdomain.MembershipRoleOwner
