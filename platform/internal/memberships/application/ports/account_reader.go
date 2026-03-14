package ports

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
)

type AccountReader interface {
	GetByID(ctx context.Context, id accdomain.AccountID) (*accdomain.Account, error)
	GetByEmail(ctx context.Context, email accdomain.Email) (*accdomain.Account, error)
}
