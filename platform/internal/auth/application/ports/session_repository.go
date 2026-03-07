package ports

import (
	"context"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(ctx context.Context, session *authdomain.RefreshSession) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*authdomain.RefreshSession, error)
	RevokeByID(ctx context.Context, id uuid.UUID) error
	ReplaceToken(ctx context.Context, sessionID uuid.UUID, newTokenHash string) error
}
