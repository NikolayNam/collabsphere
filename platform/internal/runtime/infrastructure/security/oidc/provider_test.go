package oidc

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
)

func TestBuildAuthorizationURLIncludesPKCE(t *testing.T) {
	t.Parallel()

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                 srv.URL,
			"authorization_endpoint": srv.URL + "/oauth/v2/authorize",
			"token_endpoint":         srv.URL + "/oauth/v2/token",
			"jwks_uri":               srv.URL + "/oauth/v2/keys",
		})
	}))
	defer srv.Close()

	provider := &Provider{
		issuerURL:    srv.URL,
		clientID:     "client-id",
		clientSecret: "secret",
		redirectURL:  "http://api.localhost:8080/v1/auth/zitadel/callback",
		scopes:       []string{"openid", "profile", "email"},
		httpClient:   srv.Client(),
	}

	target, err := provider.BuildAuthorizationURL(context.Background(), ports.OIDCAuthorizationRequest{
		State:               "state-token",
		Nonce:               "nonce-token",
		Prompt:              "create",
		CodeChallenge:       "challenge-value",
		CodeChallengeMethod: "S256",
	})
	if err != nil {
		t.Fatalf("BuildAuthorizationURL() error = %v", err)
	}

	parsed, err := url.Parse(target)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	query := parsed.Query()
	if got := query.Get("code_challenge"); got != "challenge-value" {
		t.Fatalf("code_challenge = %q, want challenge-value", got)
	}
	if got := query.Get("code_challenge_method"); got != "S256" {
		t.Fatalf("code_challenge_method = %q, want S256", got)
	}
}

func TestExchangeCodeSendsCodeVerifier(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	const kid = "test-key"
	nonce := "nonce-value"

	var (
		gotCode         string
		gotCodeVerifier string
	)

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"issuer":                 srv.URL,
				"authorization_endpoint": srv.URL + "/oauth/v2/authorize",
				"token_endpoint":         srv.URL + "/oauth/v2/token",
				"jwks_uri":               srv.URL + "/oauth/v2/keys",
			})
		case "/oauth/v2/token":
			if user, pass, ok := r.BasicAuth(); !ok || user != "client-id" || pass != "secret" {
				t.Fatalf("unexpected basic auth user=%q ok=%v", user, ok)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm() error = %v", err)
			}
			gotCode = r.Form.Get("code")
			gotCodeVerifier = r.Form.Get("code_verifier")
			token := signedIDToken(t, key, kid, srv.URL, "client-id", nonce)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id_token": token,
			})
		case "/oauth/v2/keys":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"keys": []map[string]any{{
					"kid": kid,
					"kty": "RSA",
					"use": "sig",
					"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
					"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes()),
				}},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	provider := &Provider{
		issuerURL:    srv.URL,
		clientID:     "client-id",
		clientSecret: "secret",
		redirectURL:  "http://api.localhost:8080/v1/auth/zitadel/callback",
		scopes:       []string{"openid", "profile", "email"},
		httpClient:   srv.Client(),
	}

	identity, err := provider.ExchangeCode(context.Background(), ports.OIDCCodeExchangeRequest{
		Code:         "auth-code",
		CodeVerifier: "verifier-token",
	})
	if err != nil {
		t.Fatalf("ExchangeCode() error = %v", err)
	}
	if identity == nil {
		t.Fatal("ExchangeCode() returned nil identity")
	}
	if gotCode != "auth-code" {
		t.Fatalf("code = %q, want auth-code", gotCode)
	}
	if gotCodeVerifier != "verifier-token" {
		t.Fatalf("code_verifier = %q, want verifier-token", gotCodeVerifier)
	}
	if identity.Nonce != nonce {
		t.Fatalf("Nonce = %q, want %q", identity.Nonce, nonce)
	}
}

func signedIDToken(t *testing.T, key *rsa.PrivateKey, kid, issuer, audience, nonce string) string {
	t.Helper()

	headerJSON, err := json.Marshal(map[string]any{
		"alg": "RS256",
		"kid": kid,
		"typ": "JWT",
	})
	if err != nil {
		t.Fatalf("Marshal header: %v", err)
	}
	payloadJSON, err := json.Marshal(map[string]any{
		"iss":            issuer,
		"sub":            "zitadel-user-1",
		"aud":            []string{audience},
		"exp":            time.Now().Add(5 * time.Minute).Unix(),
		"iat":            time.Now().Add(-time.Minute).Unix(),
		"nonce":          nonce,
		"email":          "user@example.com",
		"email_verified": true,
		"name":           "Test User",
	})
	if err != nil {
		t.Fatalf("Marshal payload: %v", err)
	}

	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := header + "." + payload
	sum := crypto.SHA256.New()
	_, _ = sum.Write([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum.Sum(nil))
	if err != nil {
		t.Fatalf("SignPKCS1v15() error = %v", err)
	}
	return strings.Join([]string{
		header,
		payload,
		base64.RawURLEncoding.EncodeToString(signature),
	}, ".")
}
