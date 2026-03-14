package ports

import (
	"context"

	memberdomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
)

type AccessAuditRepository interface {
	Append(ctx context.Context, event memberdomain.AccessAuditEvent) error
}
