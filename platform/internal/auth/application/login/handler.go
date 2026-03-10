package login

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	autherrors "github.com/NikolayNam/collabsphere/internal/auth/application/errors"
	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/google/uuid"
)

type Result struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
}

type Handler struct {
	accounts ports.AccountReader
	verifier ports.PasswordVerifier
	tokens   ports.TokenManager
	random   ports.RandomTokenGenerator
	sessions ports.SessionRepository
	clock    ports.Clock
}

func NewHandler(
	accounts ports.AccountReader,
	verifier ports.PasswordVerifier,
	tokens ports.TokenManager,
	random ports.RandomTokenGenerator,
	sessions ports.SessionRepository,
	clock ports.Clock,
) *Handler {
	return &Handler{
		accounts: accounts,
		verifier: verifier,
		tokens:   tokens,
		random:   random,
		sessions: sessions,
		clock:    clock,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	email, err := accdomain.NewEmail(cmd.Email)
	if err != nil {
		return nil, autherrors.InvalidInput("Invalid email")
	}

	acc, err := h.accounts.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, autherrors.Unauthorized("Invalid credentials")
	}

	if acc.Status() != accdomain.AccountStatusActive {
		return nil, autherrors.Forbidden("Account is not active")
	}
	if acc.PasswordHash().IsZero() {
		return nil, autherrors.Unauthorized("Invalid credentials")
	}

	if err := h.verifier.Verify(acc.PasswordHash(), cmd.Password); err != nil {
		return nil, autherrors.Unauthorized("Invalid credentials")
	}

	now := h.clock.Now()
	sessionID := uuid.New()

	refreshRaw, err := h.random.Generate()
	if err != nil {
		return nil, autherrors.Internal("generate refresh token failed", err)
	}
	refreshHash := h.random.Hash(refreshRaw)

	session, err := authdomain.NewRefreshSession(authdomain.NewRefreshSessionParams{
		ID:        sessionID,
		AccountID: acc.ID().UUID(),
		TokenHash: refreshHash,
		UserAgent: cmd.UserAgent,
		IP:        cmd.IP,
		ExpiresAt: now.Add(h.tokens.SessionTTL()),
		Now:       now,
	})
	if err != nil {
		return nil, autherrors.Internal("build refresh session failed", err)
	}

	if err := h.sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	accessExpiresAt := now.Add(h.tokens.AccessTTL())
	accessToken, err := h.tokens.GenerateAccessToken(
		ctx,
		authdomain.NewPrincipal(acc.ID().UUID(), sessionID),
		accessExpiresAt,
	)
	if err != nil {
		return nil, autherrors.Internal("generate access token failed", err)
	}

	return &Result{
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.tokens.AccessTTL().Seconds()),
	}, nil
}
