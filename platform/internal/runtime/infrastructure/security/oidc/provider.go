package oidc

import (
	"context"
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
)

type Provider struct {
	issuerURL    string
	clientID     string
	clientSecret string
	redirectURL  string
	scopes       []string
	httpClient   *http.Client
}

type discoveryDocument struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

type tokenResponse struct {
	IDToken string `json:"id_token"`
}

type idTokenHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

type audienceClaim []string

func (a *audienceClaim) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		*a = nil
		return nil
	}
	if data[0] == '[' {
		var items []string
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		*a = items
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if strings.TrimSpace(value) == "" {
		*a = nil
		return nil
	}
	*a = []string{value}
	return nil
}

type idTokenClaims struct {
	Issuer            string        `json:"iss"`
	Subject           string        `json:"sub"`
	Audience          audienceClaim `json:"aud"`
	ExpiresAt         int64         `json:"exp"`
	IssuedAt          int64         `json:"iat"`
	NotBefore         int64         `json:"nbf"`
	Nonce             string        `json:"nonce"`
	Email             string        `json:"email"`
	EmailVerified     bool          `json:"email_verified"`
	Name              string        `json:"name"`
	PreferredUsername string        `json:"preferred_username"`
}

type jwkSet struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func NewZitadelProvider(cfg config.Zitadel) (*Provider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	secret, err := cfg.ClientSecretValue()
	if err != nil {
		return nil, err
	}
	return &Provider{
		issuerURL:    strings.TrimRight(strings.TrimSpace(cfg.IssuerURL), "/"),
		clientID:     strings.TrimSpace(cfg.ClientID),
		clientSecret: secret,
		redirectURL:  strings.TrimSpace(cfg.RedirectURL),
		scopes:       cfg.ScopeList(),
		httpClient:   &http.Client{Timeout: cfg.HTTPTimeout},
	}, nil
}

func (p *Provider) Name() string {
	return "zitadel"
}

func (p *Provider) BuildAuthorizationURL(ctx context.Context, state, nonce string) (string, error) {
	if p == nil {
		return "", errors.New("oidc provider is nil")
	}
	discovery, err := p.fetchDiscovery(ctx)
	if err != nil {
		return "", err
	}
	authURL, err := url.Parse(discovery.AuthorizationEndpoint)
	if err != nil {
		return "", fmt.Errorf("parse authorization endpoint: %w", err)
	}
	query := authURL.Query()
	query.Set("client_id", p.clientID)
	query.Set("redirect_uri", p.redirectURL)
	query.Set("response_type", "code")
	query.Set("scope", strings.Join(p.scopes, " "))
	query.Set("state", state)
	query.Set("nonce", nonce)
	authURL.RawQuery = query.Encode()
	return authURL.String(), nil
}

func (p *Provider) ExchangeCode(ctx context.Context, code string) (*ports.OIDCIdentity, error) {
	if p == nil {
		return nil, errors.New("oidc provider is nil")
	}
	discovery, err := p.fetchDiscovery(ctx)
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", strings.TrimSpace(code))
	form.Set("redirect_uri", p.redirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, discovery.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	req.SetBasicAuth(p.clientID, p.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("exchange code: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	if strings.TrimSpace(tokenResp.IDToken) == "" {
		return nil, errors.New("token response does not contain id_token")
	}

	return p.verifyIDToken(ctx, discovery, tokenResp.IDToken)
}

func (p *Provider) fetchDiscovery(ctx context.Context) (*discoveryDocument, error) {
	if p == nil {
		return nil, errors.New("oidc provider is nil")
	}
	discoveryURL := p.issuerURL + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build discovery request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch discovery: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read discovery response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch discovery: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var discovery discoveryDocument
	if err := json.Unmarshal(body, &discovery); err != nil {
		return nil, fmt.Errorf("decode discovery response: %w", err)
	}
	if strings.TrimSpace(discovery.AuthorizationEndpoint) == "" || strings.TrimSpace(discovery.TokenEndpoint) == "" || strings.TrimSpace(discovery.JWKSURI) == "" {
		return nil, errors.New("discovery document is incomplete")
	}
	return &discovery, nil
}

func (p *Provider) verifyIDToken(ctx context.Context, discovery *discoveryDocument, rawToken string) (*ports.OIDCIdentity, error) {
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("id_token format is invalid")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode id_token header: %w", err)
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode id_token claims: %w", err)
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode id_token signature: %w", err)
	}

	var header idTokenHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("parse id_token header: %w", err)
	}
	if strings.TrimSpace(header.Alg) == "" || strings.TrimSpace(header.Kid) == "" {
		return nil, errors.New("id_token header is incomplete")
	}

	keys, err := p.fetchJWKS(ctx, discovery.JWKSURI)
	if err != nil {
		return nil, err
	}
	key, err := findRSAKey(keys, header.Kid)
	if err != nil {
		return nil, err
	}
	if err := verifyRSASignature(header.Alg, key, parts[0]+"."+parts[1], signature); err != nil {
		return nil, err
	}

	var claims idTokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("parse id_token claims: %w", err)
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return nil, errors.New("id_token subject is empty")
	}
	if !sameIssuer(claims.Issuer, discovery.Issuer) && !sameIssuer(claims.Issuer, p.issuerURL) {
		return nil, errors.New("id_token issuer is invalid")
	}
	if !containsAudience(claims.Audience, p.clientID) {
		return nil, errors.New("id_token audience is invalid")
	}
	now := time.Now().UTC().Unix()
	if claims.ExpiresAt <= now {
		return nil, errors.New("id_token is expired")
	}
	if claims.NotBefore != 0 && claims.NotBefore > now {
		return nil, errors.New("id_token is not active yet")
	}
	if strings.TrimSpace(claims.Nonce) == "" {
		return nil, errors.New("id_token nonce is empty")
	}

	var displayName *string
	for _, candidate := range []string{claims.Name, claims.PreferredUsername} {
		candidate = strings.TrimSpace(candidate)
		if candidate != "" {
			displayName = &candidate
			break
		}
	}

	return &ports.OIDCIdentity{
		Provider:      p.Name(),
		Subject:       claims.Subject,
		Nonce:         claims.Nonce,
		Email:         strings.TrimSpace(claims.Email),
		EmailVerified: claims.EmailVerified,
		DisplayName:   displayName,
		ClaimsJSON:    string(payloadBytes),
	}, nil
}

func (p *Provider) fetchJWKS(ctx context.Context, jwksURL string) (*jwkSet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build jwks request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read jwks response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch jwks: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var set jwkSet
	if err := json.Unmarshal(body, &set); err != nil {
		return nil, fmt.Errorf("decode jwks response: %w", err)
	}
	return &set, nil
}

func findRSAKey(set *jwkSet, kid string) (*rsa.PublicKey, error) {
	if set == nil {
		return nil, errors.New("jwks is empty")
	}
	for _, key := range set.Keys {
		if key.Kid != kid {
			continue
		}
		if key.Kty != "RSA" {
			return nil, fmt.Errorf("unsupported jwk kty %q", key.Kty)
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			return nil, fmt.Errorf("decode jwk modulus: %w", err)
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			return nil, fmt.Errorf("decode jwk exponent: %w", err)
		}
		var exponent int
		for _, b := range eBytes {
			exponent = exponent<<8 + int(b)
		}
		if exponent == 0 {
			return nil, errors.New("jwk exponent is invalid")
		}
		pub := &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: exponent}
		return pub, nil
	}
	return nil, fmt.Errorf("jwk with kid %q not found", kid)
}

func verifyRSASignature(alg string, key *rsa.PublicKey, signingInput string, signature []byte) error {
	var (
		hash crypto.Hash
		pss  bool
	)
	switch alg {
	case "RS256":
		hash = crypto.SHA256
	case "RS384":
		hash = crypto.SHA384
	case "RS512":
		hash = crypto.SHA512
	case "PS256":
		hash = crypto.SHA256
		pss = true
	case "PS384":
		hash = crypto.SHA384
		pss = true
	case "PS512":
		hash = crypto.SHA512
		pss = true
	default:
		return fmt.Errorf("unsupported id_token alg %q", alg)
	}
	h := hash.New()
	_, _ = h.Write([]byte(signingInput))
	digest := h.Sum(nil)
	if pss {
		if err := rsa.VerifyPSS(key, hash, digest, signature, nil); err != nil {
			return fmt.Errorf("verify id_token signature: %w", err)
		}
		return nil
	}
	if err := rsa.VerifyPKCS1v15(key, hash, digest, signature); err != nil {
		return fmt.Errorf("verify id_token signature: %w", err)
	}
	return nil
}

func containsAudience(audience []string, clientID string) bool {
	for _, item := range audience {
		if strings.TrimSpace(item) == clientID {
			return true
		}
	}
	return false
}

func sameIssuer(left, right string) bool {
	return strings.TrimRight(strings.TrimSpace(left), "/") == strings.TrimRight(strings.TrimSpace(right), "/")
}
