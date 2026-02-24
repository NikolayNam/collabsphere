package http

import (
	"context"
	"errors"
	"log/slog"

	appLogger "github.com/NikolayNam/collabsphere-go/internal/platform/logger"
	"github.com/danielgtaylor/huma/v2"

	userapp "github.com/NikolayNam/collabsphere-go/internal/users/application"
)

type Handler struct {
	svc *userapp.Service
}

func NewHandler(svc *userapp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateUser(ctx context.Context, input *CreateUserInput) (*UserResponse, error) {
	u, err := h.svc.CreateUser(ctx, userapp.CreateUserCmd{
		Email:     input.Body.Email,
		Password:  input.Body.Password,
		FirstName: input.Body.FirstName,
		LastName:  input.Body.LastName,
		Phone:     input.Body.Phone,
	})
	if err != nil {
		switch {
		case errors.Is(err, userapp.ErrValidation):
			return nil, huma.Error400BadRequest("Invalid input")
		case errors.Is(err, userapp.ErrConflict):
			return nil, huma.Error409Conflict("User already exists")
		default:
			l := appLogger.From(ctx)
			if l == nil {
				l = slog.Default()
			}
			l.Error(
				"create user failed",
				"err", err, // ← вот это тебе нужно
				"email", input.Body.Email,
			)
			return nil, huma.Error500InternalServerError("Internal error")
		}
	}

	resp := &UserResponse{}
	resp.Body.Email = u.Email().String()
	resp.Body.FirstName = u.FirstName()
	resp.Body.LastName = u.LastName()
	resp.Body.Phone = u.Phone()
	resp.Body.IsActive = u.IsActive()

	return resp, nil
}
