package jitsi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
)

type Manager struct {
	baseURL   string
	domain    string
	issuer    string
	audience  string
	appID     string
	appSecret []byte
}

type tokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type tokenClaims struct {
	Aud     string         `json:"aud"`
	Iss     string         `json:"iss"`
	Sub     string         `json:"sub"`
	Room    string         `json:"room"`
	Nbf     int64          `json:"nbf"`
	Exp     int64          `json:"exp"`
	Context map[string]any `json:"context,omitempty"`
}

func NewManager(cfg config.Jitsi) (*Manager, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	appSecret, err := cfg.AppSecretValue()
	if err != nil {
		return nil, err
	}
	return &Manager{
		baseURL:   strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		domain:    strings.TrimSpace(cfg.Domain),
		issuer:    cfg.IssuerValue(),
		audience:  strings.TrimSpace(cfg.Audience),
		appID:     strings.TrimSpace(cfg.AppID),
		appSecret: []byte(appSecret),
	}, nil
}

func (m *Manager) GenerateJoinToken(ctx context.Context, roomName, displayName string, moderator bool, expiresAt time.Time) (string, error) {
	_ = ctx
	if m == nil {
		return "", fmt.Errorf("jitsi manager is disabled")
	}
	if strings.TrimSpace(roomName) == "" {
		return "", fmt.Errorf("jitsi room name is required")
	}
	headerPart, err := marshalPart(tokenHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", err
	}
	claimsPart, err := marshalPart(tokenClaims{
		Aud:  m.audience,
		Iss:  m.issuer,
		Sub:  m.domain,
		Room: strings.TrimSpace(roomName),
		Nbf:  time.Now().UTC().Unix(),
		Exp:  expiresAt.UTC().Unix(),
		Context: map[string]any{
			"user": map[string]any{
				"name":      strings.TrimSpace(displayName),
				"moderator": moderator,
				"id":        m.appID,
			},
		},
	})
	if err != nil {
		return "", err
	}
	signingInput := headerPart + "." + claimsPart
	return signingInput + "." + sign(m.appSecret, signingInput), nil
}

func (m *Manager) JoinURL(roomName, token string) string {
	if m == nil || strings.TrimSpace(m.baseURL) == "" {
		return ""
	}
	roomName = strings.Trim(strings.TrimSpace(roomName), "/")
	u, err := url.Parse(m.baseURL + "/" + roomName)
	if err != nil {
		return m.baseURL
	}
	q := u.Query()
	q.Set("jwt", token)
	u.RawQuery = q.Encode()
	return u.String()
}

func marshalPart(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func sign(secret []byte, input string) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
