package logout

import (
	"context"

	autherrors "github.com/NikolayNam/collabsphere/internal/auth/application/errors"
	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
)

type Handler struct {
	sessions ports.SessionRepository
	random   ports.RandomTokenGenerator
}

func NewHandler(sessions ports.SessionRepository, random ports.RandomTokenGenerator) *Handler {
	return &Handler{sessions: sessions, random: random}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	rt, err := authdomain.NewRefreshToken(cmd.RefreshToken)
	if err != nil {
		return autherrors.RefreshTokenInvalid()
	}

	hash := h.random.Hash(rt.String())
	session, err := h.sessions.GetByTokenHash(ctx, hash)
	if err != nil {
		return err
	}
	if session == nil {
		return nil
	}

	return h.sessions.RevokeByID(ctx, session.ID())
}
