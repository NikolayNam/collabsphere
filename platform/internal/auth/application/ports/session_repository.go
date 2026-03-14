package ports

import (
	"context"
	"time"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(ctx context.Context, session *authdomain.RefreshSession) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*authdomain.RefreshSession, error)
	RotateByRefreshToken(ctx context.Context, presentedTokenHash, newTokenHash string, now time.Time) (*authdomain.RefreshSession, error)
	RevokeByID(ctx context.Context, id uuid.UUID) error
}
