package jwt

import (
	"context"
	"testing"
	"time"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/google/uuid"
)

func TestManagerGenerateAndVerifyAccessToken(t *testing.T) {
	t.Parallel()

	manager := NewManager("test-secret", 15*time.Minute, 24*time.Hour)
	principal := authdomain.NewAccountPrincipal(uuid.New(), uuid.New())

	token, err := manager.GenerateAccessToken(context.Background(), principal, time.Now().Add(15*time.Minute))
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	got, err := manager.VerifyAccessToken(context.Background(), token)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}

	if got.SubjectType != authdomain.SubjectTypeAccount {
		t.Fatalf("SubjectType = %v, want %v", got.SubjectType, authdomain.SubjectTypeAccount)
	}
	if got.AccountID != principal.AccountID {
		t.Fatalf("AccountID = %v, want %v", got.AccountID, principal.AccountID)
	}
	if got.SessionID != principal.SessionID {
		t.Fatalf("SessionID = %v, want %v", got.SessionID, principal.SessionID)
	}
	if !got.Authenticated {
		t.Fatal("Authenticated = false, want true")
	}
}

func TestManagerGenerateAndVerifyGuestAccessToken(t *testing.T) {
	t.Parallel()

	manager := NewManager("test-secret", 15*time.Minute, 24*time.Hour)
	principal := authdomain.NewGuestPrincipal(uuid.New(), uuid.New(), uuid.New())

	token, err := manager.GenerateAccessToken(context.Background(), principal, time.Now().Add(15*time.Minute))
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	got, err := manager.VerifyAccessToken(context.Background(), token)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}

	if got.SubjectType != authdomain.SubjectTypeGuest {
		t.Fatalf("SubjectType = %v, want %v", got.SubjectType, authdomain.SubjectTypeGuest)
	}
	if got.GuestID != principal.GuestID {
		t.Fatalf("GuestID = %v, want %v", got.GuestID, principal.GuestID)
	}
	if got.ChannelID != principal.ChannelID {
		t.Fatalf("ChannelID = %v, want %v", got.ChannelID, principal.ChannelID)
	}
}

func TestManagerGenerateAndVerifyServiceAccessToken(t *testing.T) {
	t.Parallel()

	manager := NewManager("test-secret", 15*time.Minute, 24*time.Hour)
	principal := authdomain.NewServicePrincipal(uuid.New(), uuid.New())

	token, err := manager.GenerateAccessToken(context.Background(), principal, time.Now().Add(15*time.Minute))
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	got, err := manager.VerifyAccessToken(context.Background(), token)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}

	if got.SubjectType != authdomain.SubjectTypeService {
		t.Fatalf("SubjectType = %v, want %v", got.SubjectType, authdomain.SubjectTypeService)
	}
	if got.ServiceID != principal.ServiceID {
		t.Fatalf("ServiceID = %v, want %v", got.ServiceID, principal.ServiceID)
	}
	if got.SessionID != principal.SessionID {
		t.Fatalf("SessionID = %v, want %v", got.SessionID, principal.SessionID)
	}
}

func TestManagerVerifyAccessTokenRejectsInvalidSignature(t *testing.T) {
	t.Parallel()

	manager := NewManager("test-secret", 15*time.Minute, 24*time.Hour)
	principal := authdomain.NewAccountPrincipal(uuid.New(), uuid.New())

	token, err := manager.GenerateAccessToken(context.Background(), principal, time.Now().Add(15*time.Minute))
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if _, err := manager.VerifyAccessToken(context.Background(), token+"broken"); err == nil {
		t.Fatal("VerifyAccessToken() error = nil, want invalid signature")
	}
}

func TestManagerVerifyAccessTokenRejectsExpiredToken(t *testing.T) {
	t.Parallel()

	manager := NewManager("test-secret", 15*time.Minute, 24*time.Hour)
	principal := authdomain.NewAccountPrincipal(uuid.New(), uuid.New())

	token, err := manager.GenerateAccessToken(context.Background(), principal, time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if _, err := manager.VerifyAccessToken(context.Background(), token); err == nil {
		t.Fatal("VerifyAccessToken() error = nil, want expired token")
	}
}
