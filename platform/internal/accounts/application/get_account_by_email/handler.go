package get_account_by_email

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/accounts/application/errors"
	"github.com/NikolayNam/collabsphere/internal/accounts/application/ports"
	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
)

type Handler struct {
	repo ports.AccountRepository
}

func NewHandler(repo ports.AccountRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*domain.Account, error) {
	email, err := domain.NewEmail(q.Email)
	if err != nil {
		return nil, errors.InvalidInput("Invalid email")
	}

	acc, err := h.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, errors.AccountNotFound()
	}

	return acc, nil
}
