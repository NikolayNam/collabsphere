package get_account_by_id

import (
	"context"
	"strings"

	"github.com/google/uuid"

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
	raw := strings.TrimSpace(q.ID)
	if raw == "" {
		return nil, errors.InvalidInput("Invalid account id")
	}

	uid, err := uuid.Parse(raw)
	if err != nil || uid == uuid.Nil {
		return nil, errors.InvalidInput("Invalid account id")
	}

	id, err := domain.AccountIDFromUUID(uid)
	if err != nil {
		return nil, errors.InvalidInput("Invalid account id")
	}

	acc, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, errors.AccountNotFound()
	}

	return acc, nil
}
