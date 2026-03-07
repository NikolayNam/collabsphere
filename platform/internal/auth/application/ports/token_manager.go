package ports

import (
	"context"
	"time"

	"github.com/NikolayNam/collabsphere/internal/auth/domain"
)

type TokenManager interface {
	GenerateAccessToken(ctx context.Context, principal domain.Principal, expiresAt time.Time) (string, error)
	VerifyAccessToken(ctx context.Context, token string) (domain.Principal, error)
	SessionTTL() time.Duration
	AccessTTL() time.Duration
}
