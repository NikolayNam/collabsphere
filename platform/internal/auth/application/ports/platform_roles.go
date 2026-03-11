package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PlatformRoleGrantRepository interface {
	EnsurePlatformRoles(ctx context.Context, accountID uuid.UUID, roles []string, grantedByAccountID *uuid.UUID, now time.Time) error
	MatchPlatformRoles(ctx context.Context, subject string, email string, emailVerified bool) ([]string, error)
}
