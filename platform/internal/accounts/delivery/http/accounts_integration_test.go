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

	"github.com/NikolayNam/collabsphere/internal/accounts/application"
	accports "github.com/NikolayNam/collabsphere/internal/accounts/application/ports"
	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	accpg "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres"
	"github.com/NikolayNam/collabsphere/internal/runtime/bootstrap"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/clock"
	"github.com/NikolayNam/collabsphere/internal/testutil/postgresitest"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestCreateAccountIntegrationSuccess(t *testing.T) {
	env := newAccountsIntegrationEnv(t, true)

	resp := env.postJSON(t, "/v1/accounts", map[string]any{
		"email":       "User@example.com",
		"password":    "Secret123",
		"displayName": "Alice",
	})
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}

	var body struct {
		ID          uuid.UUID `json:"id"`
		Email       string    `json:"email"`
		DisplayName *string   `json:"displayName"`
		IsActive    bool      `json:"isActive"`
	}
	decodeJSON(t, resp.Body, &body)

	if body.Email != "user@example.com" {
		t.Fatalf("email = %q, want user@example.com", body.Email)
	}
	if body.DisplayName == nil || *body.DisplayName != "Alice" {
		t.Fatalf("displayName = %v, want Alice", body.DisplayName)
	}
	if !body.IsActive {
		t.Fatal("isActive = false, want true")
	}

	var storedEmail string
	if err := env.queryRowContext(t, `SELECT email FROM iam.accounts WHERE id = $1`, body.ID).Scan(&storedEmail); err != nil {
		t.Fatalf("query account: %v", err)
	}
	if storedEmail != "user@example.com" {
		t.Fatalf("stored email = %q, want user@example.com", storedEmail)
	}

	var passwordHash string
	if err := env.queryRowContext(t, `SELECT password_hash FROM auth.password_credentials WHERE account_id = $1`, body.ID).Scan(&passwordHash); err != nil {
		t.Fatalf("query password credential: %v", err)
	}
	if strings.TrimSpace(passwordHash) == "" {
		t.Fatal("password_hash is empty")
	}
	if passwordHash == "Secret123" {
		t.Fatal("password_hash must not equal raw password")
	}
}

func TestCreateAccountIntegrationDuplicateEmail(t *testing.T) {
	env := newAccountsIntegrationEnv(t, true)

	first := env.postJSON(t, "/v1/accounts", map[string]any{
		"email":    "duplicate@example.com",
		"password": "Secret123",
	})
	first.Body.Close()
	if first.StatusCode != nethttp.StatusCreated {
		t.Fatalf("first status = %d, want 201", first.StatusCode)
	}

	second := env.postJSON(t, "/v1/accounts", map[string]any{
		"email":    "duplicate@example.com",
		"password": "Secret123",
	})
	defer second.Body.Close()
	if second.StatusCode != nethttp.StatusConflict {
		t.Fatalf("second status = %d, want 409", second.StatusCode)
	}
}

func TestCreateAccountIntegrationInvalidEmail(t *testing.T) {
	env := newAccountsIntegrationEnv(t, true)

	resp := env.postJSON(t, "/v1/accounts", map[string]any{
		"email":    "not-an-email",
		"password": "Secret123",
	})
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestCreateAccountIntegrationWeakPassword(t *testing.T) {
	env := newAccountsIntegrationEnv(t, true)

	resp := env.postJSON(t, "/v1/accounts", map[string]any{
		"email":    "weak@example.com",
		"password": "short",
	})
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestCreateAccountIntegrationDisabledWhenLocalSignupOff(t *testing.T) {
	env := newAccountsIntegrationEnv(t, false)

	resp := env.postJSON(t, "/v1/accounts", map[string]any{
		"email":    "blocked@example.com",
		"password": "Secret123",
	})
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
}

type accountsIntegrationEnv struct {
	server  *httptest.Server
	queryDB *sql.DB
	adminDB *sql.DB
	appDB   *sql.DB
	dbName  string
}

func newAccountsIntegrationEnv(t *testing.T, localSignupEnabled bool) *accountsIntegrationEnv {
	t.Helper()

	testDB := postgresitest.NewTempDatabase(t, "collabsphere_accounts_it")
	postgresitest.ApplyBundledMigrations(t, testDB.QueryDB)

	conf := postgresitest.TestConfig(testDB.ConnConfig, "app")
	conf.Auth.LocalSignupEnabled = localSignupEnabled
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db := bootstrap.MustOpenGormDB(conf, logger)
	bootstrap.RegisterDBHooks(db)

	repo := accpg.NewAccountRepo(db)
	service := application.New(repo, testPasswordHasher{}, clock.NewSystemClock(), nil, "")
	handler := NewHandler(service, localSignupEnabled)

	root := chi.NewRouter()
	apiV1 := chi.NewRouter()
	root.Mount("/v1", apiV1)
	api := bootstrap.NewAPI(apiV1, conf)
	Register(api, handler, nil)

	server := httptest.NewServer(root)
	appDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}

	env := &accountsIntegrationEnv{
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

type testPasswordHasher struct{}

func (testPasswordHasher) Hash(raw string) (accdomain.PasswordHash, error) {
	return accdomain.NewPasswordHash("hashed:" + raw)
}

var _ accports.PasswordHasher = testPasswordHasher{}

func (e *accountsIntegrationEnv) postJSON(t *testing.T, path string, payload map[string]any) *nethttp.Response {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	resp, err := nethttp.Post(e.server.URL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http.Post: %v", err)
	}
	return resp
}

func (e *accountsIntegrationEnv) queryRowContext(t *testing.T, query string, args ...any) *sql.Row {
	t.Helper()
	return e.queryDB.QueryRowContext(context.Background(), query, args...)
}

func decodeJSON(t *testing.T, r io.Reader, target any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(target); err != nil {
		t.Fatalf("json.Decode: %v", err)
	}
}
