package authcallback

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterServesCallbackPage(t *testing.T) {
	router := chi.NewRouter()
	Register(router, "CollabSphere")

	req := httptest.NewRequest(http.MethodGet, "/auth/callback", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "/v1/auth/exchange") {
		t.Fatalf("expected callback page to include exchange endpoint")
	}
	if !strings.Contains(body, "/v1/auth/zitadel/login?return_to=/auth/callback") {
		t.Fatalf("expected callback page to include login link")
	}
	if !strings.Contains(body, "Готово к входу") {
		t.Fatalf("expected default status to be rendered server-side")
	}
}

func TestRegisterRendersErrorStateServerSide(t *testing.T) {
	router := chi.NewRouter()
	Register(router, "CollabSphere")

	req := httptest.NewRequest(http.MethodGet, "/auth/callback?error=access_denied&error_description=Verified+email+is+required+for+first+external+login", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Вход завершился ошибкой") {
		t.Fatalf("expected server-rendered error title")
	}
	if !strings.Contains(body, "Verified email is required for first external login") {
		t.Fatalf("expected error description to be rendered")
	}
	if !strings.Contains(body, "Подтвердите email в ZITADEL") {
		t.Fatalf("expected recovery hint for unverified email")
	}
}
