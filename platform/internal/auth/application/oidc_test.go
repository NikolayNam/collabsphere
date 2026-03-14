package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	sharedtx "github.com/NikolayNam/collabsphere/shared/tx"
	"github.com/google/uuid"
)

type typedNilOIDCProvider struct{}

func (p *typedNilOIDCProvider) Name() string { return "zitadel" }

func (p *typedNilOIDCProvider) BuildAuthorizationURL(ctx context.Context, req ports.OIDCAuthorizationRequest) (string, error) {
	panic("should not be called")
}

func (p *typedNilOIDCProvider) ExchangeCode(ctx context.Context, req ports.OIDCCodeExchangeRequest) (*ports.OIDCIdentity, error) {
	panic("should not be called")
}

func TestHasOIDCProviderRejectsTypedNil(t *testing.T) {
	var provider *typedNilOIDCProvider
	if hasOIDCProvider(provider) {
		t.Fatal("typed nil provider must be treated as unavailable")
	}
}

type fakeTxManager struct{}

func (fakeTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

var _ sharedtx.Manager = fakeTxManager{}

type fakeOIDCProvider struct {
	authReq      ports.OIDCAuthorizationRequest
	exchangeReq  ports.OIDCCodeExchangeRequest
	authURL      string
	exchangeResp *ports.OIDCIdentity
	exchangeErr  error
}

func (f *fakeOIDCProvider) Name() string { return "zitadel" }

func (f *fakeOIDCProvider) BuildAuthorizationURL(ctx context.Context, req ports.OIDCAuthorizationRequest) (string, error) {
	f.authReq = req
	if f.authURL == "" {
		return "https://issuer.example/authorize", nil
	}
	return f.authURL, nil
}

func (f *fakeOIDCProvider) ExchangeCode(ctx context.Context, req ports.OIDCCodeExchangeRequest) (*ports.OIDCIdentity, error) {
	f.exchangeReq = req
	if f.exchangeErr != nil {
		return nil, f.exchangeErr
	}
	return f.exchangeResp, nil
}

type fakeOIDCStateRepo struct {
	state *ports.OAuthStateRecord
	nonce *ports.OIDCNonceRecord
}

func (f *fakeOIDCStateRepo) CreateState(ctx context.Context, record *ports.OAuthStateRecord) error {
	if record == nil {
		return errors.New("state is nil")
	}
	copied := *record
	f.state = &copied
	return nil
}

func (f *fakeOIDCStateRepo) CreateNonce(ctx context.Context, record *ports.OIDCNonceRecord) error {
	if record == nil {
		return errors.New("nonce is nil")
	}
	copied := *record
	f.nonce = &copied
	return nil
}

func (f *fakeOIDCStateRepo) GetStateByHash(ctx context.Context, provider, stateHash string) (*ports.OAuthStateRecord, error) {
	if f.state == nil || f.state.Provider != provider || f.state.StateHash != stateHash {
		return nil, nil
	}
	copied := *f.state
	return &copied, nil
}

func (f *fakeOIDCStateRepo) GetNonceByStateID(ctx context.Context, provider string, stateID uuid.UUID) (*ports.OIDCNonceRecord, error) {
	if f.nonce == nil || f.nonce.Provider != provider || f.nonce.OAuthStateID != stateID {
		return nil, nil
	}
	copied := *f.nonce
	return &copied, nil
}

func (f *fakeOIDCStateRepo) MarkStateUsed(ctx context.Context, id uuid.UUID, at time.Time) error {
	if f.state != nil && f.state.ID == id {
		value := at
		f.state.UsedAt = &value
		f.state.UpdatedAt = &value
	}
	return nil
}

func (f *fakeOIDCStateRepo) MarkNonceUsed(ctx context.Context, id uuid.UUID, at time.Time) error {
	if f.nonce != nil && f.nonce.ID == id {
		value := at
		f.nonce.UsedAt = &value
		f.nonce.UpdatedAt = &value
	}
	return nil
}

type fakeRandomTokenGenerator struct {
	values []string
	index  int
}

func (f *fakeRandomTokenGenerator) Generate() (string, error) {
	if f.index >= len(f.values) {
		return "", errors.New("no more random values")
	}
	value := f.values[f.index]
	f.index++
	return value, nil
}

func (f *fakeRandomTokenGenerator) Hash(raw string) string {
	return "hash:" + raw
}

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time { return f.now }

func TestPKCEChallengeS256(t *testing.T) {
	got, err := pkceChallengeS256("verifier")
	if err != nil {
		t.Fatalf("pkceChallengeS256() error = %v", err)
	}
	if got != "iMnq5o6zALKXGivsnlom_0F5_WYda32GHkxlV7mq7hQ" {
		t.Fatalf("pkceChallengeS256() = %q", got)
	}
}

func TestBeginOIDCLoginStoresCodeVerifierAndPassesPKCEChallenge(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	states := &fakeOIDCStateRepo{}
	provider := &fakeOIDCProvider{}
	random := &fakeRandomTokenGenerator{values: []string{"state-token", "nonce-token", "verifier-token"}}
	flow := newOIDCFlow(
		fakeTxManager{},
		nil,
		nil,
		states,
		nil,
		nil,
		OIDCPlatformAutoGrantPolicy{},
		provider,
		nil,
		random,
		nil,
		fakeClock{now: now},
		15*time.Minute,
		15*time.Minute,
		time.Minute,
	)

	res, err := flow.BeginLogin(context.Background(), BeginOIDCLoginCmd{
		ReturnTo: "http://collabsphere.localhost:3002/auth/callback",
		Intent:   "login",
	})
	if err != nil {
		t.Fatalf("BeginLogin() error = %v", err)
	}
	if res == nil || res.AuthorizationURL == "" {
		t.Fatal("BeginLogin() must return authorization URL")
	}
	if states.state == nil {
		t.Fatal("CreateState() was not called")
	}
	if states.state.CodeVerifier != "verifier-token" {
		t.Fatalf("stored CodeVerifier = %q, want verifier-token", states.state.CodeVerifier)
	}
	if provider.authReq.CodeChallengeMethod != "S256" {
		t.Fatalf("CodeChallengeMethod = %q, want S256", provider.authReq.CodeChallengeMethod)
	}
	wantChallenge, err := pkceChallengeS256("verifier-token")
	if err != nil {
		t.Fatalf("pkceChallengeS256() error = %v", err)
	}
	if provider.authReq.CodeChallenge != wantChallenge {
		t.Fatalf("CodeChallenge = %q, want %q", provider.authReq.CodeChallenge, wantChallenge)
	}
	if provider.authReq.State != "state-token" {
		t.Fatalf("State = %q, want state-token", provider.authReq.State)
	}
	if provider.authReq.Nonce != "nonce-token" {
		t.Fatalf("Nonce = %q, want nonce-token", provider.authReq.Nonce)
	}
}
