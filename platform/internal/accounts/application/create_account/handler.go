package create_account

import (
    "context"
    stdErrors "errors"

    "github.com/NikolayNam/collabsphere/internal/accounts/application/errors"
    "github.com/NikolayNam/collabsphere/internal/accounts/application/ports"
    "github.com/NikolayNam/collabsphere/internal/accounts/application/validation"
    "github.com/NikolayNam/collabsphere/internal/accounts/domain"
)

type Handler struct {
    repo   ports.AccountRepository
    hasher ports.PasswordHasher
    clock  ports.Clock
}

func NewHandler(repo ports.AccountRepository, hasher ports.PasswordHasher, clock ports.Clock) *Handler {
    return &Handler{repo: repo, hasher: hasher, clock: clock}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*domain.Account, error) {
    email, err := domain.NewEmail(cmd.Email)
    if err != nil {
        return nil, errors.InvalidInput("Invalid email")
    }

    if err := validation.ValidatePassword(cmd.Password); err != nil {
        return nil, err
    }

    passwordHash, err := h.hasher.Hash(cmd.Password)
    if err != nil {
        return nil, errors.Internal("hash password failed", err)
    }

    acc, err := domain.NewAccount(domain.NewAccountParams{
        ID:           domain.NewAccountID(),
        Email:        email,
        PasswordHash: passwordHash,
        DisplayName:  cmd.DisplayName,
        Now:          h.clock.Now(),
    })
    if err != nil {
        return nil, errors.InvalidInput("Invalid account data")
    }

    if err := h.repo.Create(ctx, acc); err != nil {
        if stdErrors.Is(err, errors.ErrConflict) {
            return nil, errors.AccountAlreadyExists()
        }
        return nil, err
    }
    return acc, nil
}
