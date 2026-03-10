package ports

import (
	"context"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	authports "github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	"github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"github.com/google/uuid"
)

type RoleBindingRepository interface {
	ListRoles(ctx context.Context, accountID uuid.UUID) ([]domain.Role, error)
	ListAccountIDsByRole(ctx context.Context, role domain.Role) ([]uuid.UUID, error)
	ReplaceRoles(ctx context.Context, accountID uuid.UUID, roles []domain.Role, grantedByAccountID *uuid.UUID, now time.Time) error
}

type AuditRepository interface {
	Append(ctx context.Context, event domain.AuditEvent) error
}

type AccountReader interface {
	GetByID(ctx context.Context, id accdomain.AccountID) (*accdomain.Account, error)
}

type DashboardReader interface {
	GetDashboardSummary(ctx context.Context) (*domain.DashboardSummary, error)
}

type UploadQueueReader interface {
	ListUploadQueue(ctx context.Context, query domain.UploadQueueQuery) ([]domain.UploadQueueItem, int, error)
}

type ZitadelAdminClient = authports.ZitadelAdminClient
