package jwt

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/google/uuid"
)

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	sessionTTL time.Duration
}

type accessTokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type accessTokenClaims struct {
	SubjectType string `json:"st"`
	Subject     string `json:"sub"`
	SessionID   string `json:"sid"`
	ChannelID   string `json:"cid,omitempty"`
	IssuedAt    int64  `json:"iat"`
	ExpiresAt   int64  `json:"exp"`
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

	if len(m.secret) == 0 {
		return "", errors.New("jwt secret is empty")
	}
	if !principal.Authenticated || principal.SessionID == uuid.Nil {
		return "", errors.New("principal is not authenticated")
	}
	if expiresAt.IsZero() {
		return "", errors.New("access token expiry is required")
	}

	claims, err := buildClaims(principal, expiresAt)
	if err != nil {
		return "", err
	}

	headerPart, err := marshalTokenPart(accessTokenHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", fmt.Errorf("marshal jwt header: %w", err)
	}
	claimsPart, err := marshalTokenPart(claims)
	if err != nil {
		return "", fmt.Errorf("marshal jwt claims: %w", err)
	}

	signingInput := headerPart + "." + claimsPart
	signature := m.sign(signingInput)
	return signingInput + "." + signature, nil
}

func (m *Manager) VerifyAccessToken(ctx context.Context, token string) (authdomain.Principal, error) {
	_ = ctx

	if len(m.secret) == 0 {
		return authdomain.Principal{}, errors.New("jwt secret is empty")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return authdomain.Principal{}, errors.New("jwt token format is invalid")
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSignature := m.sign(signingInput)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return authdomain.Principal{}, errors.New("jwt signature is invalid")
	}

	var header accessTokenHeader
	if err := unmarshalTokenPart(parts[0], &header); err != nil {
		return authdomain.Principal{}, fmt.Errorf("decode jwt header: %w", err)
	}
	if header.Alg != "HS256" || header.Typ != "JWT" {
		return authdomain.Principal{}, errors.New("jwt header is invalid")
	}

	var claims accessTokenClaims
	if err := unmarshalTokenPart(parts[1], &claims); err != nil {
		return authdomain.Principal{}, fmt.Errorf("decode jwt claims: %w", err)
	}
	if claims.ExpiresAt <= time.Now().UTC().Unix() {
		return authdomain.Principal{}, errors.New("jwt token is expired")
	}

	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil || sessionID == uuid.Nil {
		return authdomain.Principal{}, errors.New("jwt session is invalid")
	}

	switch authdomain.SubjectType(claims.SubjectType) {
	case authdomain.SubjectTypeAccount:
		accountID, err := uuid.Parse(claims.Subject)
		if err != nil || accountID == uuid.Nil {
			return authdomain.Principal{}, errors.New("jwt subject is invalid")
		}
		return authdomain.NewAccountPrincipal(accountID, sessionID), nil
	case authdomain.SubjectTypeGuest:
		guestID, err := uuid.Parse(claims.Subject)
		if err != nil || guestID == uuid.Nil {
			return authdomain.Principal{}, errors.New("jwt subject is invalid")
		}
		channelID, err := uuid.Parse(claims.ChannelID)
		if err != nil || channelID == uuid.Nil {
			return authdomain.Principal{}, errors.New("jwt guest channel is invalid")
		}
		return authdomain.NewGuestPrincipal(guestID, sessionID, channelID), nil
	case authdomain.SubjectTypeService:
		serviceID, err := uuid.Parse(claims.Subject)
		if err != nil || serviceID == uuid.Nil {
			return authdomain.Principal{}, errors.New("jwt subject is invalid")
		}
		return authdomain.NewServicePrincipal(serviceID, sessionID), nil
	default:
		return authdomain.Principal{}, errors.New("jwt subject type is invalid")
	}
}

func buildClaims(principal authdomain.Principal, expiresAt time.Time) (accessTokenClaims, error) {
	claims := accessTokenClaims{
		SubjectType: string(principal.SubjectType),
		SessionID:   principal.SessionID.String(),
		IssuedAt:    time.Now().UTC().Unix(),
		ExpiresAt:   expiresAt.UTC().Unix(),
	}

	switch {
	case principal.IsAccount():
		claims.Subject = principal.AccountID.String()
	case principal.IsGuest():
		if principal.ChannelID == uuid.Nil {
			return accessTokenClaims{}, errors.New("guest principal channel is required")
		}
		claims.Subject = principal.GuestID.String()
		claims.ChannelID = principal.ChannelID.String()
	case principal.IsService():
		claims.Subject = principal.ServiceID.String()
	default:
		return accessTokenClaims{}, errors.New("principal is not authenticated")
	}

	return claims, nil
}

func marshalTokenPart(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func unmarshalTokenPart(part string, target any) error {
	data, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func (m *Manager) sign(input string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
