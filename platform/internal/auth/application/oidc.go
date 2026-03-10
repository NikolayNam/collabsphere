package application

import (
	"context"
	"reflect"
	"strings"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	autherrors "github.com/NikolayNam/collabsphere/internal/auth/application/errors"
	"github.com/NikolayNam/collabsphere/internal/auth/application/login"
	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/NikolayNam/collabsphere/shared/tx"
	"github.com/google/uuid"
)

type BeginOIDCLoginCmd struct{}

type BeginOIDCLoginResult struct {
	AuthorizationURL string
}

type CompleteOIDCCallbackCmd struct {
	State     string
	Code      string
	UserAgent *string
	IP        *string
}

type oidcFlow struct {
	tx                 tx.Manager
	accounts           ports.AccountReader
	externalIdentities ports.ExternalIdentityRepository
	states             ports.OIDCStateRepository
	provider           ports.OIDCProvider
	tokens             ports.TokenManager
	random             ports.RandomTokenGenerator
	sessions           ports.SessionRepository
	clock              ports.Clock
	stateTTL           time.Duration
	nonceTTL           time.Duration
}

func newOIDCFlow(
	txm tx.Manager,
	accounts ports.AccountReader,
	externalIdentities ports.ExternalIdentityRepository,
	states ports.OIDCStateRepository,
	provider ports.OIDCProvider,
	tokens ports.TokenManager,
	random ports.RandomTokenGenerator,
	sessions ports.SessionRepository,
	clock ports.Clock,
	stateTTL time.Duration,
	nonceTTL time.Duration,
) *oidcFlow {
	return &oidcFlow{
		tx:                 txm,
		accounts:           accounts,
		externalIdentities: externalIdentities,
		states:             states,
		provider:           provider,
		tokens:             tokens,
		random:             random,
		sessions:           sessions,
		clock:              clock,
		stateTTL:           stateTTL,
		nonceTTL:           nonceTTL,
	}
}

func (f *oidcFlow) BeginLogin(ctx context.Context, _ BeginOIDCLoginCmd) (*BeginOIDCLoginResult, error) {
	if f == nil || !hasOIDCProvider(f.provider) || f.states == nil || f.random == nil || f.clock == nil || f.tx == nil {
		return nil, autherrors.Unavailable("OIDC login is unavailable")
	}

	stateRaw, err := f.random.Generate()
	if err != nil {
		return nil, autherrors.Internal("generate oidc state failed", err)
	}
	nonceRaw, err := f.random.Generate()
	if err != nil {
		return nil, autherrors.Internal("generate oidc nonce failed", err)
	}

	now := f.clock.Now()
	stateRecord := &ports.OAuthStateRecord{
		ID:        uuid.New(),
		Provider:  f.provider.Name(),
		StateHash: f.random.Hash(stateRaw),
		ExpiresAt: now.Add(f.stateTTL),
		CreatedAt: now,
	}
	nonceRecord := &ports.OIDCNonceRecord{
		ID:           uuid.New(),
		Provider:     f.provider.Name(),
		OAuthStateID: stateRecord.ID,
		NonceHash:    f.random.Hash(nonceRaw),
		ExpiresAt:    now.Add(f.nonceTTL),
		CreatedAt:    now,
	}

	if err := f.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := f.states.CreateState(ctx, stateRecord); err != nil {
			return err
		}
		return f.states.CreateNonce(ctx, nonceRecord)
	}); err != nil {
		return nil, err
	}

	authorizationURL, err := f.provider.BuildAuthorizationURL(ctx, stateRaw, nonceRaw)
	if err != nil {
		return nil, autherrors.Internal("build oidc authorization url failed", err)
	}
	return &BeginOIDCLoginResult{AuthorizationURL: authorizationURL}, nil
}

func (f *oidcFlow) CompleteCallback(ctx context.Context, cmd CompleteOIDCCallbackCmd) (*login.Result, error) {
	if f == nil || !hasOIDCProvider(f.provider) || f.states == nil || f.externalIdentities == nil || f.accounts == nil || f.sessions == nil || f.tokens == nil || f.random == nil || f.clock == nil || f.tx == nil {
		return nil, autherrors.Unavailable("OIDC login is unavailable")
	}
	if strings.TrimSpace(cmd.State) == "" || strings.TrimSpace(cmd.Code) == "" {
		return nil, autherrors.InvalidInput("OIDC callback state and code are required")
	}

	stateHash := f.random.Hash(strings.TrimSpace(cmd.State))
	state, err := f.states.GetStateByHash(ctx, f.provider.Name(), stateHash)
	if err != nil {
		return nil, err
	}
	if state == nil || state.UsedAt != nil || !state.ExpiresAt.After(f.clock.Now()) {
		return nil, autherrors.Unauthorized("OIDC callback is invalid or expired")
	}

	nonce, err := f.states.GetNonceByStateID(ctx, f.provider.Name(), state.ID)
	if err != nil {
		return nil, err
	}
	if nonce == nil || nonce.UsedAt != nil || !nonce.ExpiresAt.After(f.clock.Now()) {
		return nil, autherrors.Unauthorized("OIDC callback is invalid or expired")
	}

	identity, err := f.provider.ExchangeCode(ctx, cmd.Code)
	if err != nil {
		return nil, autherrors.Unauthorized("OIDC code exchange failed")
	}
	if identity == nil || strings.TrimSpace(identity.Subject) == "" {
		return nil, autherrors.Unauthorized("OIDC identity is invalid")
	}
	if f.random.Hash(identity.Nonce) != nonce.NonceHash {
		return nil, autherrors.Unauthorized("OIDC nonce is invalid")
	}

	var result *login.Result
	err = f.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		now := f.clock.Now()
		account, externalIdentity, resolveErr := f.resolveAccount(ctx, identity, now)
		if resolveErr != nil {
			return resolveErr
		}
		if account.Status() != accdomain.AccountStatusActive {
			return autherrors.Forbidden("Account is not active")
		}
		if externalIdentity != nil {
			if err := f.externalIdentities.TouchLogin(ctx, externalIdentity.ID, externalIdentity.Email, externalIdentity.EmailVerified, externalIdentity.DisplayName, externalIdentity.ClaimsJSON, now); err != nil {
				return err
			}
		}

		sessionID := uuid.New()
		refreshRaw, err := f.random.Generate()
		if err != nil {
			return autherrors.Internal("generate refresh token failed", err)
		}
		session, err := authdomain.NewRefreshSession(authdomain.NewRefreshSessionParams{
			ID:        sessionID,
			AccountID: account.ID().UUID(),
			TokenHash: f.random.Hash(refreshRaw),
			UserAgent: cmd.UserAgent,
			IP:        cmd.IP,
			ExpiresAt: now.Add(f.tokens.SessionTTL()),
			Now:       now,
		})
		if err != nil {
			return autherrors.Internal("build refresh session failed", err)
		}
		if err := f.sessions.Create(ctx, session); err != nil {
			return err
		}

		accessToken, err := f.tokens.GenerateAccessToken(ctx, authdomain.NewPrincipal(account.ID().UUID(), sessionID), now.Add(f.tokens.AccessTTL()))
		if err != nil {
			return autherrors.Internal("generate access token failed", err)
		}
		if err := f.states.MarkStateUsed(ctx, state.ID, now); err != nil {
			return err
		}
		if err := f.states.MarkNonceUsed(ctx, nonce.ID, now); err != nil {
			return err
		}
		result = &login.Result{
			AccessToken:  accessToken,
			RefreshToken: refreshRaw,
			TokenType:    "Bearer",
			ExpiresIn:    int64(f.tokens.AccessTTL().Seconds()),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (f *oidcFlow) resolveAccount(ctx context.Context, identity *ports.OIDCIdentity, now time.Time) (*accdomain.Account, *ports.ExternalIdentityRecord, error) {
	externalIdentity, err := f.externalIdentities.GetByProviderSubject(ctx, f.provider.Name(), identity.Subject)
	if err != nil {
		return nil, nil, err
	}
	if externalIdentity != nil {
		accountID, err := accdomain.AccountIDFromUUID(externalIdentity.AccountID)
		if err != nil {
			return nil, nil, autherrors.Internal("invalid external identity account id", err)
		}
		account, err := f.accounts.GetByID(ctx, accountID)
		if err != nil {
			return nil, nil, err
		}
		if account == nil {
			return nil, nil, autherrors.Unauthorized("External identity is not linked to an account")
		}
		externalIdentity.Email = normalizeOptional(identity.Email)
		externalIdentity.EmailVerified = identity.EmailVerified
		externalIdentity.DisplayName = identity.DisplayName
		externalIdentity.ClaimsJSON = identity.ClaimsJSON
		return account, externalIdentity, nil
	}

	if strings.TrimSpace(identity.Email) == "" || !identity.EmailVerified {
		return nil, nil, autherrors.Forbidden("Verified email is required for first external login")
	}
	email, err := accdomain.NewEmail(identity.Email)
	if err != nil {
		return nil, nil, autherrors.Internal("external identity email is invalid", err)
	}

	account, err := f.accounts.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if account == nil {
		account, err = accdomain.NewAccount(accdomain.NewAccountParams{
			ID:           accdomain.NewAccountID(),
			Email:        email,
			DisplayName:  identity.DisplayName,
			PasswordHash: "",
			Now:          now,
		})
		if err != nil {
			return nil, nil, autherrors.Internal("build external account failed", err)
		}
		if err := f.accounts.Create(ctx, account); err != nil {
			return nil, nil, err
		}
	}

	externalIdentity = &ports.ExternalIdentityRecord{
		ID:              uuid.New(),
		Provider:        f.provider.Name(),
		ExternalSubject: identity.Subject,
		AccountID:       account.ID().UUID(),
		Email:           normalizeOptional(identity.Email),
		EmailVerified:   identity.EmailVerified,
		DisplayName:     identity.DisplayName,
		ClaimsJSON:      identity.ClaimsJSON,
		LastLoginAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       &now,
	}
	if err := f.externalIdentities.Create(ctx, externalIdentity); err != nil {
		return nil, nil, err
	}
	return account, externalIdentity, nil
}

func hasOIDCProvider(provider ports.OIDCProvider) bool {
	if provider == nil {
		return false
	}
	value := reflect.ValueOf(provider)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return !value.IsNil()
	default:
		return true
	}
}

func normalizeOptional(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
