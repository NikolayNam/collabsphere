package httpbind

import (
	"context"
	"errors"
	"testing"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/google/uuid"
)

func TestRequireAccountIDRejectsGuest(t *testing.T) {
	ctx := middleware.WithPrincipal(context.Background(), authdomain.NewGuestPrincipal(uuid.New(), uuid.New(), uuid.New()))
	want := errors.New("unauthorized")

	_, err := RequireAccountID(ctx, want)
	if !errors.Is(err, want) {
		t.Fatalf("RequireAccountID() error = %v, want %v", err, want)
	}
}

func TestParseUUIDTrimsSpace(t *testing.T) {
	id := uuid.New()
	parsed, err := ParseUUID("  "+id.String()+"  ", errors.New("invalid"))
	if err != nil {
		t.Fatalf("ParseUUID() error = %v", err)
	}
	if parsed != id {
		t.Fatalf("ParseUUID() = %s, want %s", parsed, id)
	}
}
