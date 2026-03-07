package http

import (
	"context"
	"net/http"

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

	return mapper.ToAccountResponse(u, http.StatusCreated), nil
}

func (h *Handler) GetAccountById(ctx context.Context, input *dto.GetAccountByIdInput) (*dto.AccountResponse, error) {
	u, err := h.svc.GetAccountById(ctx, application.GetAccountByIdQuery{
		ID: input.ID,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}

	return mapper.ToAccountResponse(u, http.StatusOK), nil
}

func (h *Handler) GetAccountByEmail(ctx context.Context, input *dto.GetAccountByEmailInput) (*dto.AccountResponse, error) {
	u, err := h.svc.GetAccountByEmail(ctx, application.GetAccountByEmailQuery{
		Email: input.Email,
	})

	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return mapper.ToAccountResponse(u, http.StatusOK), nil
}
