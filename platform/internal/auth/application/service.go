package application

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/auth/application/errors"
	"github.com/NikolayNam/collabsphere/internal/auth/application/login"
	"github.com/NikolayNam/collabsphere/internal/auth/application/logout"
	"github.com/NikolayNam/collabsphere/internal/auth/application/me"
	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	"github.com/NikolayNam/collabsphere/internal/auth/application/refresh"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
)

var (
	ErrValidation   = errors.ErrValidation
	ErrUnauthorized = errors.ErrUnauthorized
	ErrForbidden    = errors.ErrForbidden
)

type LoginCmd = login.Command
type RefreshCmd = refresh.Command
type LogoutCmd = logout.Command
type MeQuery = me.Query

type Service struct {
	login   *login.Handler
	refresh *refresh.Handler
	logout  *logout.Handler
	me      *me.Handler
}

func New(
	accounts ports.AccountReader,
	verifier ports.PasswordVerifier,
	tokens ports.TokenManager,
	random ports.RandomTokenGenerator,
	sessions ports.SessionRepository,
	clock ports.Clock,
) *Service {
	return &Service{
		login:   login.NewHandler(accounts, verifier, tokens, random, sessions, clock),
		refresh: refresh.NewHandler(accounts, sessions, tokens, random, clock),
		logout:  logout.NewHandler(sessions, random),
		me:      me.NewHandler(accounts),
	}
}

func (s *Service) Login(ctx context.Context, cmd LoginCmd) (*login.Result, error) {
	return s.login.Handle(ctx, cmd)
}

func (s *Service) Refresh(ctx context.Context, cmd RefreshCmd) (*refresh.Result, error) {
	return s.refresh.Handle(ctx, cmd)
}

func (s *Service) Logout(ctx context.Context, cmd LogoutCmd) error {
	return s.logout.Handle(ctx, cmd)
}

func (s *Service) Me(ctx context.Context, principal authdomain.Principal) (*accdomain.Account, error) {
	return s.me.Handle(ctx, me.Query{Principal: principal})
}
