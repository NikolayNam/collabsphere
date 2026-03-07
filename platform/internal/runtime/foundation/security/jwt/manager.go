package jwt

import (
	"context"
	"errors"
	"time"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
)

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	sessionTTL time.Duration
}

func NewManager(secret string, accessTTL, sessionTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		sessionTTL: sessionTTL,
	}
}

func (m *Manager) AccessTTL() time.Duration {
	return m.accessTTL
}

func (m *Manager) SessionTTL() time.Duration {
	return m.sessionTTL
}

func (m *Manager) GenerateAccessToken(ctx context.Context, principal authdomain.Principal, expiresAt time.Time) (string, error) {
	_ = ctx
	_ = principal
	_ = expiresAt

	// Здесь либо github.com/golang-jwt/jwt/v5,
	// либо другой проверенный пакет.
	// Пока skeleton:
	return "", errors.New("not implemented")
}

func (m *Manager) VerifyAccessToken(ctx context.Context, token string) (authdomain.Principal, error) {
	_ = ctx
	_ = token

	return authdomain.Principal{}, errors.New("not implemented")
}
