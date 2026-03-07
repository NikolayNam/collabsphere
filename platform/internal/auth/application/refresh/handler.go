package refresh

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	autherrors "github.com/NikolayNam/collabsphere/internal/auth/application/errors"
	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
)

type Result struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
}

type Handler struct {
	accounts ports.AccountReader
	sessions ports.SessionRepository
	tokens   ports.TokenManager
	random   ports.RandomTokenGenerator
	clock    ports.Clock
}

func NewHandler(
	accounts ports.AccountReader,
	sessions ports.SessionRepository,
	tokens ports.TokenManager,
	random ports.RandomTokenGenerator,
	clock ports.Clock,
) *Handler {
	return &Handler{
		accounts: accounts,
		sessions: sessions,
		tokens:   tokens,
		random:   random,
		clock:    clock,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	rt, err := authdomain.NewRefreshToken(cmd.RefreshToken)
	if err != nil {
		return nil, autherrors.RefreshTokenInvalid()
	}

	hash := h.random.Hash(rt.String())
	session, err := h.sessions.GetByTokenHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, autherrors.RefreshTokenInvalid()
	}

	now := h.clock.Now()
	if session.IsRevoked() || session.IsExpired(now) {
		return nil, autherrors.RefreshTokenInvalid()
	}

	accID, err := accdomain.AccountIDFromUUID(session.AccountID())
	if err != nil {
		return nil, autherrors.RefreshTokenInvalid()
	}

	acc, err := h.accounts.GetByID(ctx, accID)
	if err != nil {
		return nil, err
	}
	if acc == nil || acc.Status() != accdomain.AccountStatusActive {
		return nil, autherrors.RefreshTokenInvalid()
	}

	newRefreshRaw, err := h.random.Generate()
	if err != nil {
		return nil, autherrors.Internal("generate refresh token failed", err)
	}
	newRefreshHash := h.random.Hash(newRefreshRaw)

	if err := h.sessions.ReplaceToken(ctx, session.ID(), newRefreshHash); err != nil {
		return nil, err
	}

	accessExpiresAt := now.Add(h.tokens.AccessTTL())
	accessToken, err := h.tokens.GenerateAccessToken(
		ctx,
		authdomain.NewPrincipal(acc.ID().UUID(), session.ID()),
		accessExpiresAt,
	)
	if err != nil {
		return nil, autherrors.Internal("generate access token failed", err)
	}

	return &Result{
		AccessToken:  accessToken,
		RefreshToken: newRefreshRaw,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.tokens.AccessTTL().Seconds()),
	}, nil
}
