package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestAccessLogSkipsQuietPaths(t *testing.T) {
	var out bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&out, nil))

	handler := AccessLog(logger, AccessLogOptions{
		QuietPaths: []string{"/health"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no log output for quiet path, got %q", out.String())
	}
}

func TestAccessLogWritesRegularRequests(t *testing.T) {
	var out bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&out, nil))

	handler := AccessLog(logger, AccessLogOptions{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/accounts", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	got := out.String()
	if !strings.Contains(got, "\"path\":\"/v1/accounts\"") {
		t.Fatalf("expected log to contain request path, got %q", got)
	}
	if !strings.Contains(got, "\"status\":201") {
		t.Fatalf("expected log to contain response status, got %q", got)
	}
	if !strings.Contains(got, "\"route\":\"unmatched\"") {
		t.Fatalf("expected log to contain fallback route, got %q", got)
	}
}

func TestAccessLogWritesRoutePattern(t *testing.T) {
	var out bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&out, nil))

	router := chi.NewRouter()
	router.Use(AccessLog(logger, AccessLogOptions{}))
	router.Get("/v1/organizations/{organizationId}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/organizations/123", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	got := out.String()
	if !strings.Contains(got, "\"route\":\"/v1/organizations/{organizationId}\"") {
		t.Fatalf("expected log to contain route pattern, got %q", got)
	}
}
