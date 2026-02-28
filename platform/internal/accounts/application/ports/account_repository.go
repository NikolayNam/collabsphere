package ports

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
)

type AccountRepository interface {
	Create(ctx context.Context, account *domain.Account) error
	GetByID(ctx context.Context, id domain.AccountID) (*domain.Account, error)
	GetByEmail(ctx context.Context, email domain.Email) (*domain.Account, error)
	ExistsByEmail(ctx context.Context, email domain.Email) (bool, error)
}
