//go:build integration

package http

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	accountsapp "github.com/NikolayNam/collabsphere/internal/accounts/application"
	accountshttp "github.com/NikolayNam/collabsphere/internal/accounts/delivery/http"
	accpg "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
	authapp "github.com/NikolayNam/collabsphere/internal/auth/application"
	authpg "github.com/NikolayNam/collabsphere/internal/auth/repository/postgres"
	platformpg "github.com/NikolayNam/collabsphere/internal/platformops/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/bootstrap"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/bcrypt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/jwt"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/security/tokens"
	dbtx "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
	"github.com/NikolayNam/collabsphere/internal/testutil/postgresitest"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestLegacyPasswordLoginIntegrationFlow(t *testing.T) {
	env := newAuthIntegrationEnv(t, true)

	account := env.createAccount(t, "user@example.com", "Secret123", "Alice")

	loginResp := env.postJSON(t, "/v1/auth/login", map[string]any{
		"email":    "user@example.com",
		"password": "Secret123",
	}, "")
	defer loginResp.Body.Close()
	if loginResp.StatusCode != nethttp.StatusOK {
		t.Fatalf("login status = %d, want 200", loginResp.StatusCode)
	}

	var loginBody tokenPayload
	decodeJSONAuth(t, loginResp.Body, &loginBody)
	if strings.TrimSpace(loginBody.AccessToken) == "" {
		t.Fatal("accessToken is empty")
	}
	if strings.TrimSpace(loginBody.RefreshToken) == "" {
		t.Fatal("refreshToken is empty")
	}
	if loginBody.TokenType != "Bearer" {
		t.Fatalf("tokenType = %q, want Bearer", loginBody.TokenType)
	}

	meResp := env.get(t, "/v1/auth/me", loginBody.AccessToken)
	defer meResp.Body.Close()
	if meResp.StatusCode != nethttp.StatusOK {
		t.Fatalf("me status = %d, want 200", meResp.StatusCode)
	}

	var meBody struct {
		ID    uuid.UUID `json:"id"`
		Email string    `json:"email"`
	}
	decodeJSONAuth(t, meResp.Body, &meBody)
	if meBody.ID != account.ID {
		t.Fatalf("me.id = %s, want %s", meBody.ID, account.ID)
	}
	if meBody.Email != "user@example.com" {
		t.Fatalf("me.email = %q, want user@example.com", meBody.Email)
	}

	refreshResp := env.postJSON(t, "/v1/auth/refresh", map[string]any{
		"refreshToken": loginBody.RefreshToken,
	}, "")
	defer refreshResp.Body.Close()
	if refreshResp.StatusCode != nethttp.StatusOK {
		t.Fatalf("refresh status = %d, want 200", refreshResp.StatusCode)
	}

	var refreshBody tokenPayload
	decodeJSONAuth(t, refreshResp.Body, &refreshBody)
	if refreshBody.RefreshToken == loginBody.RefreshToken {
		t.Fatal("refresh token must rotate")
	}

	logoutResp := env.postJSON(t, "/v1/auth/logout", map[string]any{
		"refreshToken": refreshBody.RefreshToken,
	}, "")
	defer logoutResp.Body.Close()
	if logoutResp.StatusCode != nethttp.StatusNoContent {
		t.Fatalf("logout status = %d, want 204", logoutResp.StatusCode)
	}

	reuseResp := env.postJSON(t, "/v1/auth/refresh", map[string]any{
		"refreshToken": refreshBody.RefreshToken,
	}, "")
	defer reuseResp.Body.Close()
	if reuseResp.StatusCode != nethttp.StatusUnauthorized {
		t.Fatalf("post-logout refresh status = %d, want 401", reuseResp.StatusCode)
	}

	var reuseProblem problemPayload
	decodeJSONAuth(t, reuseResp.Body, &reuseProblem)
	if reuseProblem.Code != "AUTH_REFRESH_INVALID" {
		t.Fatalf("post-logout refresh code = %q, want AUTH_REFRESH_INVALID", reuseProblem.Code)
	}
}

func TestLegacyPasswordLoginIntegrationRejectsReuseAfterRotation(t *testing.T) {
	env := newAuthIntegrationEnv(t, true)
	env.createAccount(t, "reuse@example.com", "Secret123", "Reuse")

	loginResp := env.postJSON(t, "/v1/auth/login", map[string]any{
		"email":    "reuse@example.com",
		"password": "Secret123",
	}, "")
	defer loginResp.Body.Close()
	if loginResp.StatusCode != nethttp.StatusOK {
		t.Fatalf("login status = %d, want 200", loginResp.StatusCode)
	}

	var loginBody tokenPayload
	decodeJSONAuth(t, loginResp.Body, &loginBody)

	firstRefreshResp := env.postJSON(t, "/v1/auth/refresh", map[string]any{
		"refreshToken": loginBody.RefreshToken,
	}, "")
	defer firstRefreshResp.Body.Close()
	if firstRefreshResp.StatusCode != nethttp.StatusOK {
		t.Fatalf("first refresh status = %d, want 200", firstRefreshResp.StatusCode)
	}

	var firstRefresh tokenPayload
	decodeJSONAuth(t, firstRefreshResp.Body, &firstRefresh)

	reuseResp := env.postJSON(t, "/v1/auth/refresh", map[string]any{
		"refreshToken": loginBody.RefreshToken,
	}, "")
	defer reuseResp.Body.Close()
	if reuseResp.StatusCode != nethttp.StatusUnauthorized {
		t.Fatalf("reuse status = %d, want 401", reuseResp.StatusCode)
	}

	var reuseProblem problemPayload
	decodeJSONAuth(t, reuseResp.Body, &reuseProblem)
	if reuseProblem.Code != "AUTH_REFRESH_INVALID" {
		t.Fatalf("reuse code = %q, want AUTH_REFRESH_INVALID", reuseProblem.Code)
	}

	revokedResp := env.postJSON(t, "/v1/auth/refresh", map[string]any{
		"refreshToken": firstRefresh.RefreshToken,
	}, "")
	defer revokedResp.Body.Close()
	if revokedResp.StatusCode != nethttp.StatusUnauthorized {
		t.Fatalf("revoked status = %d, want 401", revokedResp.StatusCode)
	}
}

func TestLegacyPasswordLoginIntegrationDisabledByFlag(t *testing.T) {
	env := newAuthIntegrationEnv(t, false)
	env.createAccount(t, "disabled@example.com", "Secret123", "Disabled")

	resp := env.postJSON(t, "/v1/auth/login", map[string]any{
		"email":    "disabled@example.com",
		"password": "Secret123",
	}, "")
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}

	var problem problemPayload
	decodeJSONAuth(t, resp.Body, &problem)
	if problem.Code != "PASSWORD_LOGIN_DISABLED" {
		t.Fatalf("code = %q, want PASSWORD_LOGIN_DISABLED", problem.Code)
	}
}

func TestAuthMeIntegrationRequiresAuthentication(t *testing.T) {
	env := newAuthIntegrationEnv(t, true)

	resp := env.get(t, "/v1/auth/me", "")
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}

	var problem problemPayload
	decodeJSONAuth(t, resp.Body, &problem)
	if problem.Code != "AUTH_UNAUTHORIZED" {
		t.Fatalf("code = %q, want AUTH_UNAUTHORIZED", problem.Code)
	}
}

type authIntegrationEnv struct {
	server  *httptest.Server
	queryDB *sql.DB
	adminDB *sql.DB
	appDB   *sql.DB
	dbName  string
}

type createdAccount struct {
	ID    uuid.UUID
	Email string
}

type tokenPayload struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type problemPayload struct {
	Status int    `json:"status"`
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

func newAuthIntegrationEnv(t *testing.T, passwordLoginEnabled bool) *authIntegrationEnv {
	t.Helper()

	testDB := postgresitest.NewTempDatabase(t, "collabsphere_auth_it")
	postgresitest.ApplyBundledMigrations(t, testDB.QueryDB)

	conf := postgresitest.TestConfig(testDB.ConnConfig, "app")
	conf.Auth.PasswordLoginEnabled = passwordLoginEnabled
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db := bootstrap.MustOpenGormDB(conf, logger)
	bootstrap.RegisterDBHooks(db)

	accountRepo := accpg.NewAccountRepo(db)
	passwordHasher := bcrypt.NewBcryptHasher()
	clk := clock.NewSystemClock()
	tokenGen := tokens.NewGenerator()

	accountService := accountsapp.New(accountRepo, passwordHasher, clk, nil, "")
	accountHandler := accountshttp.NewHandler(accountService, conf.Auth.LocalSignupEnabled)

	jwtManager := jwt.NewManager(conf.Auth.JWTSecret, conf.Auth.AccessTTL, conf.Auth.RefreshSessionTTL)
	authService := authapp.New(
		accountRepo,
		passwordHasher,
		jwtManager,
		tokenGen,
		authpg.NewSessionRepo(db),
		clk,
		dbtx.New(db),
		authpg.NewExternalIdentityRepo(db),
		authpg.NewOIDCStateRepo(db),
		authpg.NewOneTimeCodeRepo(db),
		platformpg.NewRepo(db),
		authapp.OIDCPlatformAutoGrantPolicy{},
		nil,
		nil,
		conf.Auth.Zitadel.StateTTL,
		conf.Auth.Zitadel.NonceTTL,
		conf.Auth.BrowserTicketTTL,
	)
	authHandler := NewHandler(authService, passwordLoginEnabled, false, BrowserFlowConfig{})

	root := chi.NewRouter()
	apiV1 := chi.NewRouter()
	root.Mount("/v1", apiV1)
	api := bootstrap.NewAPI(apiV1, conf)
	accountshttp.Register(api, accountHandler, jwtManager)
	Register(root, api, authHandler, jwtManager)

	server := httptest.NewServer(root)
	appDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}

	env := &authIntegrationEnv{
		server:  server,
		queryDB: testDB.QueryDB,
		adminDB: testDB.AdminDB,
		appDB:   appDB,
		dbName:  testDB.DBName,
	}
	t.Cleanup(func() {
		server.Close()
		_ = appDB.Close()
	})
	return env
}

func (e *authIntegrationEnv) createAccount(t *testing.T, email, password, displayName string) createdAccount {
	t.Helper()

	resp := e.postJSON(t, "/v1/accounts", map[string]any{
		"email":       email,
		"password":    password,
		"displayName": displayName,
	}, "")
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusCreated {
		t.Fatalf("create account status = %d, want 201", resp.StatusCode)
	}

	var body struct {
		ID    uuid.UUID `json:"id"`
		Email string    `json:"email"`
	}
	decodeJSONAuth(t, resp.Body, &body)
	return createdAccount{ID: body.ID, Email: body.Email}
}

func (e *authIntegrationEnv) postJSON(t *testing.T, path string, payload map[string]any, accessToken string) *nethttp.Response {
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

func (e *authIntegrationEnv) get(t *testing.T, path, accessToken string) *nethttp.Response {
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

func decodeJSONAuth(t *testing.T, r io.Reader, target any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(target); err != nil {
		t.Fatalf("json.Decode: %v", err)
	}
}
