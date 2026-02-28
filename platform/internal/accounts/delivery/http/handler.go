package http

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"

	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"

	"github.com/NikolayNam/collabsphere/internal/accounts/application"
	"github.com/NikolayNam/collabsphere/internal/accounts/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/accounts/delivery/http/mapper"
)

type Handler struct {
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateAccount(ctx context.Context, input *dto.CreateAccountInput) (*dto.AccountResponse, error) {
	u, err := h.svc.CreateAccount(ctx, application.CreateAccountCmd{
		Email:     input.Body.Email,
		Password:  input.Body.Password,
		FirstName: input.Body.FirstName,
		LastName:  input.Body.LastName,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	return mapper.ToAccountResponse(u), nil
}

func (h *Handler) GetAccountById(ctx context.Context, input *dto.GetAccountByIdInput) (*dto.AccountResponse, error) {
	u, err := h.svc.GetAccountById(ctx, application.GetAccountByIdQuery{
		ID: input.ID,
	})
	if err != nil {
		switch {
		case errors.Is(err, application.ErrValidation):
			return nil, huma.Error400BadRequest("Invalid account id")
		case errors.Is(err, application.ErrNotFound):
			return nil, huma.Error404NotFound("Account not found")
		default:
			// h.logInternal(ctx, "get account failed", "err", err, "id", input.ID)
			return nil, huma.Error500InternalServerError("Internal error")
		}
	}

	return mapper.ToAccountResponse(u), nil
}

func (h *Handler) GetAccountByEmail(ctx context.Context, input *dto.GetAccountByEmailInput) (*dto.AccountResponse, error) {
	u, err := h.svc.GetAccountByEmail(ctx, application.GetAccountByEmailQuery{
		Email: input.Email,
	})

	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToAccountResponse(u), nil
}
