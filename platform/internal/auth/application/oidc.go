package application

import (
	"context"
	"errors"
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

const oidcBrowserExchangePurpose = "oidc_browser_exchange"

type BeginOIDCLoginCmd struct {
	ReturnTo string
	Intent   string
}

type BeginOIDCLoginResult struct {
	AuthorizationURL string
}

type ResolveOIDCCallbackStateCmd struct {
	State string
}

type ResolveOIDCCallbackStateResult struct {
	ReturnTo string
	Intent   string
}

type CompleteOIDCCallbackCmd struct {
	State     string
	Code      string
	UserAgent *string
	IP        *string
}

type CompleteOIDCCallbackResult struct {
	ReturnTo       string
	ExchangeTicket string
}

type ExchangeAuthTicketCmd struct {
	Ticket    string
	UserAgent *string
	IP        *string
}

type ExchangeAuthTicketResult struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
	Provider     string
	Intent       string
	IsNewAccount bool
}

type oidcFlow struct {
	tx                 tx.Manager
	accounts           ports.AccountReader
	externalIdentities ports.ExternalIdentityRepository
	states             ports.OIDCStateRepository
	oneTimeCodes       ports.OneTimeCodeRepository
	platformRoles      ports.PlatformRoleGrantRepository
	autoGrantPolicy    oidcPlatformAutoGrantPolicy
	provider           ports.OIDCProvider
	tokens             ports.TokenManager
	random             ports.RandomTokenGenerator
	sessions           ports.SessionRepository
	clock              ports.Clock
	stateTTL           time.Duration
	nonceTTL           time.Duration
	browserTicketTTL   time.Duration
}

func newOIDCFlow(
	txm tx.Manager,
	accounts ports.AccountReader,
	externalIdentities ports.ExternalIdentityRepository,
	states ports.OIDCStateRepository,
	oneTimeCodes ports.OneTimeCodeRepository,
	platformRoles ports.PlatformRoleGrantRepository,
	autoGrantPolicy OIDCPlatformAutoGrantPolicy,
	provider ports.OIDCProvider,
	tokens ports.TokenManager,
	random ports.RandomTokenGenerator,
	sessions ports.SessionRepository,
	clock ports.Clock,
	stateTTL time.Duration,
	nonceTTL time.Duration,
	browserTicketTTL time.Duration,
) *oidcFlow {
	return &oidcFlow{
		tx:                 txm,
		accounts:           accounts,
		externalIdentities: externalIdentities,
		states:             states,
		oneTimeCodes:       oneTimeCodes,
		platformRoles:      platformRoles,
		autoGrantPolicy:    newOIDCPlatformAutoGrantPolicy(autoGrantPolicy),
		provider:           provider,
		tokens:             tokens,
		random:             random,
		sessions:           sessions,
		clock:              clock,
		stateTTL:           stateTTL,
		nonceTTL:           nonceTTL,
		browserTicketTTL:   browserTicketTTL,
	}
}

func (f *oidcFlow) BeginLogin(ctx context.Context, cmd BeginOIDCLoginCmd) (*BeginOIDCLoginResult, error) {
	if f == nil || !hasOIDCProvider(f.provider) || f.states == nil || f.random == nil || f.clock == nil || f.tx == nil {
		return nil, autherrors.Unavailable("OIDC login is unavailable")
	}
	returnTo := strings.TrimSpace(cmd.ReturnTo)
	if returnTo == "" {
		return nil, autherrors.InvalidInput("OIDC return URL is required")
	}
	intent := normalizeOIDCIntent(cmd.Intent)
	if intent == "" {
		return nil, autherrors.InvalidInput("OIDC intent is invalid")
	}

	stateRaw, err := f.random.Generate()
	if err != nil {
		return nil, autherrors.Internal("generate oidc state failed", err)
	}
	nonceRaw, err := f.random.Generate()
	if err != nil {
		return nil, autherrors.Internal("generate oidc nonce failed", err)
	}
	codeVerifier, err := f.random.Generate()
	if err != nil {
		return nil, autherrors.Internal("generate oidc pkce verifier failed", err)
	}
	codeChallenge, err := pkceChallengeS256(codeVerifier)
	if err != nil {
		return nil, autherrors.Internal("build oidc pkce challenge failed", err)
	}

	now := f.clock.Now()
	stateRecord := &ports.OAuthStateRecord{
		ID:           uuid.New(),
		Provider:     f.provider.Name(),
		StateHash:    f.random.Hash(stateRaw),
		CodeVerifier: codeVerifier,
		ReturnTo:     returnTo,
		Intent:       intent,
		ExpiresAt:    now.Add(f.stateTTL),
		CreatedAt:    now,
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

	authorizationURL, err := f.provider.BuildAuthorizationURL(ctx, ports.OIDCAuthorizationRequest{
		State:               stateRaw,
		Nonce:               nonceRaw,
		Prompt:              oidcPromptForIntent(intent),
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: "S256",
	})
	if err != nil {
		return nil, autherrors.Internal("build oidc authorization url failed", err)
	}
	return &BeginOIDCLoginResult{AuthorizationURL: authorizationURL}, nil
}

func (f *oidcFlow) ResolveCallbackState(ctx context.Context, cmd ResolveOIDCCallbackStateCmd) (*ResolveOIDCCallbackStateResult, error) {
	state, err := f.getStateByRaw(ctx, cmd.State)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, nil
	}
	return &ResolveOIDCCallbackStateResult{ReturnTo: state.ReturnTo, Intent: state.Intent}, nil
}

func (f *oidcFlow) CompleteCallback(ctx context.Context, cmd CompleteOIDCCallbackCmd) (*CompleteOIDCCallbackResult, error) {
	if f == nil || !hasOIDCProvider(f.provider) || f.states == nil || f.externalIdentities == nil || f.accounts == nil || f.oneTimeCodes == nil || f.random == nil || f.clock == nil || f.tx == nil {
		return nil, autherrors.Unavailable("OIDC login is unavailable")
	}
	if strings.TrimSpace(cmd.State) == "" || strings.TrimSpace(cmd.Code) == "" {
		return nil, autherrors.InvalidInput("OIDC callback state and code are required")
	}

	state, err := f.getStateByRaw(ctx, cmd.State)
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

	if strings.TrimSpace(state.CodeVerifier) == "" {
		return nil, autherrors.Internal("oidc pkce verifier is missing", errors.New("oauth state code verifier is empty"))
	}

	identity, err := f.provider.ExchangeCode(ctx, ports.OIDCCodeExchangeRequest{
		Code:         cmd.Code,
		CodeVerifier: state.CodeVerifier,
	})
	if err != nil {
		return nil, autherrors.Unauthorized("OIDC code exchange failed")
	}
	if identity == nil || strings.TrimSpace(identity.Subject) == "" {
		return nil, autherrors.Unauthorized("OIDC identity is invalid")
	}
	if f.random.Hash(identity.Nonce) != nonce.NonceHash {
		return nil, autherrors.Unauthorized("OIDC nonce is invalid")
	}

	var result *CompleteOIDCCallbackResult
	err = f.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		now := f.clock.Now()
		account, externalIdentity, isNewAccount, resolveErr := f.resolveAccount(ctx, identity, now)
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
		if err := f.autoGrantPlatformRoles(ctx, account.ID().UUID(), identity, now); err != nil {
			return autherrors.Internal("auto grant platform admin failed", err)
		}
		ticketRaw, err := f.random.Generate()
		if err != nil {
			return autherrors.Internal("generate exchange ticket failed", err)
		}
		code := &ports.OneTimeCodeRecord{
			ID:           uuid.New(),
			Purpose:      oidcBrowserExchangePurpose,
			CodeHash:     f.random.Hash(ticketRaw),
			AccountID:    account.ID().UUID(),
			Provider:     f.provider.Name(),
			Intent:       state.Intent,
			IsNewAccount: isNewAccount,
			ExpiresAt:    now.Add(f.browserTicketTTL),
			CreatedAt:    now,
		}
		if err := f.oneTimeCodes.Create(ctx, code); err != nil {
			return err
		}
		if err := f.states.MarkStateUsed(ctx, state.ID, now); err != nil {
			return err
		}
		if err := f.states.MarkNonceUsed(ctx, nonce.ID, now); err != nil {
			return err
		}
		result = &CompleteOIDCCallbackResult{
			ReturnTo:       state.ReturnTo,
			ExchangeTicket: ticketRaw,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (f *oidcFlow) ExchangeTicket(ctx context.Context, cmd ExchangeAuthTicketCmd) (*ExchangeAuthTicketResult, error) {
	if f == nil || f.oneTimeCodes == nil || f.accounts == nil || f.sessions == nil || f.tokens == nil || f.random == nil || f.clock == nil || f.tx == nil {
		return nil, autherrors.Unavailable("OIDC login is unavailable")
	}
	if strings.TrimSpace(cmd.Ticket) == "" {
		return nil, autherrors.InvalidInput("Exchange ticket is required")
	}

	ticketHash := f.random.Hash(strings.TrimSpace(cmd.Ticket))
	record, err := f.oneTimeCodes.GetByCodeHash(ctx, oidcBrowserExchangePurpose, ticketHash)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, autherrors.Unauthorized("Exchange ticket is invalid")
	}

	var result *ExchangeAuthTicketResult
	err = f.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		now := f.clock.Now()
		lockedRecord, err := f.oneTimeCodes.GetByCodeHash(ctx, oidcBrowserExchangePurpose, ticketHash)
		if err != nil {
			return err
		}
		if lockedRecord == nil || lockedRecord.UsedAt != nil || !lockedRecord.ExpiresAt.After(now) {
			return autherrors.Unauthorized("Exchange ticket is invalid or expired")
		}
		marked, err := f.oneTimeCodes.MarkUsed(ctx, lockedRecord.ID, now)
		if err != nil {
			return err
		}
		if !marked {
			return autherrors.Unauthorized("Exchange ticket is invalid or already used")
		}

		accountID, err := accdomain.AccountIDFromUUID(lockedRecord.AccountID)
		if err != nil {
			return autherrors.Unauthorized("Exchange ticket is invalid")
		}
		account, err := f.accounts.GetByID(ctx, accountID)
		if err != nil {
			return err
		}
		if account == nil || account.Status() != accdomain.AccountStatusActive {
			return autherrors.Forbidden("Account is not active")
		}

		loginResult, err := f.issueTokens(ctx, account.ID().UUID(), cmd.UserAgent, cmd.IP, now)
		if err != nil {
			return err
		}
		result = &ExchangeAuthTicketResult{
			AccessToken:  loginResult.AccessToken,
			RefreshToken: loginResult.RefreshToken,
			TokenType:    loginResult.TokenType,
			ExpiresIn:    loginResult.ExpiresIn,
			Provider:     lockedRecord.Provider,
			Intent:       lockedRecord.Intent,
			IsNewAccount: lockedRecord.IsNewAccount,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (f *oidcFlow) resolveAccount(ctx context.Context, identity *ports.OIDCIdentity, now time.Time) (*accdomain.Account, *ports.ExternalIdentityRecord, bool, error) {
	externalIdentity, err := f.externalIdentities.GetByProviderSubject(ctx, f.provider.Name(), identity.Subject)
	if err != nil {
		return nil, nil, false, err
	}
	if externalIdentity != nil {
		accountID, err := accdomain.AccountIDFromUUID(externalIdentity.AccountID)
		if err != nil {
			return nil, nil, false, autherrors.Internal("invalid external identity account id", err)
		}
		account, err := f.accounts.GetByID(ctx, accountID)
		if err != nil {
			return nil, nil, false, err
		}
		if account == nil {
			return nil, nil, false, autherrors.Unauthorized("External identity is not linked to an account")
		}
		externalIdentity.Email = normalizeOptional(identity.Email)
		externalIdentity.EmailVerified = identity.EmailVerified
		externalIdentity.DisplayName = identity.DisplayName
		externalIdentity.ClaimsJSON = identity.ClaimsJSON
		return account, externalIdentity, false, nil
	}

	if strings.TrimSpace(identity.Email) == "" || !identity.EmailVerified {
		return nil, nil, false, autherrors.Forbidden("Verified email is required for first external login")
	}
	email, err := accdomain.NewEmail(identity.Email)
	if err != nil {
		return nil, nil, false, autherrors.Internal("external identity email is invalid", err)
	}

	account, err := f.accounts.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, false, err
	}
	isNewAccount := false
	if account == nil {
		account, err = accdomain.NewAccount(accdomain.NewAccountParams{
			ID:           accdomain.NewAccountID(),
			Email:        email,
			DisplayName:  identity.DisplayName,
			PasswordHash: "",
			Now:          now,
		})
		if err != nil {
			return nil, nil, false, autherrors.Internal("build external account failed", err)
		}
		if err := f.accounts.Create(ctx, account); err != nil {
			return nil, nil, false, err
		}
		isNewAccount = true
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
		return nil, nil, false, err
	}
	return account, externalIdentity, isNewAccount, nil
}

func (f *oidcFlow) issueTokens(ctx context.Context, accountID uuid.UUID, userAgent, ip *string, now time.Time) (*login.Result, error) {
	sessionID := uuid.New()
	refreshRaw, err := f.random.Generate()
	if err != nil {
		return nil, autherrors.Internal("generate refresh token failed", err)
	}
	session, err := authdomain.NewRefreshSession(authdomain.NewRefreshSessionParams{
		ID:        sessionID,
		AccountID: accountID,
		TokenHash: f.random.Hash(refreshRaw),
		UserAgent: userAgent,
		IP:        ip,
		ExpiresAt: now.Add(f.tokens.SessionTTL()),
		Now:       now,
	})
	if err != nil {
		return nil, autherrors.Internal("build refresh session failed", err)
	}
	if err := f.sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	accessToken, err := f.tokens.GenerateAccessToken(ctx, authdomain.NewPrincipal(accountID, sessionID), now.Add(f.tokens.AccessTTL()))
	if err != nil {
		return nil, autherrors.Internal("generate access token failed", err)
	}
	return &login.Result{
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
		TokenType:    "Bearer",
		ExpiresIn:    int64(f.tokens.AccessTTL().Seconds()),
	}, nil
}

func (f *oidcFlow) getStateByRaw(ctx context.Context, rawState string) (*ports.OAuthStateRecord, error) {
	if f == nil || f.states == nil || f.random == nil || !hasOIDCProvider(f.provider) {
		return nil, autherrors.Unavailable("OIDC login is unavailable")
	}
	stateHash := f.random.Hash(strings.TrimSpace(rawState))
	return f.states.GetStateByHash(ctx, f.provider.Name(), stateHash)
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

func normalizeOIDCIntent(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "login":
		return "login"
	case "signup":
		return "signup"
	default:
		return ""
	}
}

func oidcPromptForIntent(intent string) string {
	if intent == "signup" {
		return "create"
	}
	return ""
}
