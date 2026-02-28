package application

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/accounts/application/create_account"
	"github.com/NikolayNam/collabsphere/internal/accounts/application/errors"
	"github.com/NikolayNam/collabsphere/internal/accounts/application/get_account_by_email"
	"github.com/NikolayNam/collabsphere/internal/accounts/application/get_account_by_id"
	"github.com/NikolayNam/collabsphere/internal/accounts/application/ports"
	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
)

var (
	ErrValidation = errors.ErrValidation
	ErrNotFound   = errors.ErrNotFound
)

type CreateAccountCmd = create_account.Command
type GetAccountByIdQuery = get_account_by_id.Query
type GetAccountByEmailQuery = get_account_by_email.Query

type Service struct {
	create     *create_account.Handler
	getById    *get_account_by_id.Handler
	getByEmail *get_account_by_email.Handler
}

func New(repo ports.AccountRepository, hasher ports.PasswordHasher, clock ports.Clock) *Service {
	return &Service{
		create:     create_account.NewHandler(repo, hasher, clock),
		getById:    get_account_by_id.NewHandler(repo),
		getByEmail: get_account_by_email.NewHandler(repo),
	}
}

func (s *Service) CreateAccount(ctx context.Context, cmd CreateAccountCmd) (*domain.Account, error) {
	return s.create.Handle(ctx, cmd)
}

func (s *Service) GetAccountById(ctx context.Context, q GetAccountByIdQuery) (*domain.Account, error) {
	return s.getById.Handle(ctx, q)
}

func (s *Service) GetAccountByEmail(ctx context.Context, q GetAccountByEmailQuery) (*domain.Account, error) {
	return s.getByEmail.Handle(ctx, q)
}
