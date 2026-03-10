package ports

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
)

type AccountReader interface {
	GetByEmail(ctx context.Context, email accdomain.Email) (*accdomain.Account, error)
	GetByID(ctx context.Context, id accdomain.AccountID) (*accdomain.Account, error)
	Create(ctx context.Context, account *accdomain.Account) error
}
