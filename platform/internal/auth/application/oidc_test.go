package application

import (
	"context"
	"testing"

	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
)

type typedNilOIDCProvider struct{}

func (p *typedNilOIDCProvider) Name() string { return "zitadel" }

func (p *typedNilOIDCProvider) BuildAuthorizationURL(ctx context.Context, state, nonce string) (string, error) {
	panic("should not be called")
}

func (p *typedNilOIDCProvider) ExchangeCode(ctx context.Context, code string) (*ports.OIDCIdentity, error) {
	panic("should not be called")
}

func TestHasOIDCProviderRejectsTypedNil(t *testing.T) {
	var provider *typedNilOIDCProvider
	if hasOIDCProvider(provider) {
		t.Fatal("typed nil provider must be treated as unavailable")
	}
}
