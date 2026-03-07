package me

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	autherrors "github.com/NikolayNam/collabsphere/internal/auth/application/errors"
	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
)

type Handler struct {
	accounts ports.AccountReader
}

func NewHandler(accounts ports.AccountReader) *Handler {
	return &Handler{accounts: accounts}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*accdomain.Account, error) {
	if !q.Principal.Authenticated {
		return nil, autherrors.Unauthorized("Authentication required")
	}

	accID, err := accdomain.AccountIDFromUUID(q.Principal.AccountID)
	if err != nil {
		return nil, autherrors.Unauthorized("Authentication required")
	}

	acc, err := h.accounts.GetByID(ctx, accID)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, autherrors.Unauthorized("Authentication required")
	}

	return acc, nil
}
