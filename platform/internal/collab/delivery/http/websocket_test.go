package http

import (
	"context"
	"net/http/httptest"
	"testing"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/google/uuid"
)

type verifierStub struct {
	principal authdomain.Principal
}

func (v verifierStub) VerifyAccessToken(ctx context.Context, token string) (authdomain.Principal, error) {
	return v.principal, nil
}

func TestAuthenticateWSPrincipalRejectsQueryTokenByDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/collab?access_token=token", nil)
	got := authenticateWSPrincipal(req, verifierStub{principal: authdomain.NewAccountPrincipal(uuid.New(), uuid.New())}, false)
	if got.Authenticated {
		t.Fatal("expected anonymous principal when query token is disabled")
	}
}

func TestAuthenticateWSPrincipalAllowsQueryTokenWhenEnabled(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/collab?access_token=token", nil)
	got := authenticateWSPrincipal(req, verifierStub{principal: authdomain.NewAccountPrincipal(uuid.New(), uuid.New())}, true)
	if !got.Authenticated || !got.IsAccount() {
		t.Fatalf("expected authenticated account principal, got %#v", got)
	}
}
