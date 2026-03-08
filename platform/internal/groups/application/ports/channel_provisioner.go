package ports

import (
	"context"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/groups/domain"
)

type ChannelProvisioner interface {
	ProvisionDefaults(ctx context.Context, groupID domain.GroupID, ownerAccountID accdomain.AccountID, now time.Time) error
}
