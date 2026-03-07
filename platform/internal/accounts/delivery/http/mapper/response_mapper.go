package mapper

import (
    "github.com/google/uuid"

    "github.com/NikolayNam/collabsphere/internal/accounts/delivery/http/dto"
    "github.com/NikolayNam/collabsphere/internal/accounts/domain"
)

func ToAccountResponse(a *domain.Account, status int) *dto.AccountResponse {
    if a == nil {
        return nil
    }

    return &dto.AccountResponse{
        Status: status,
        Body: struct {
            ID          uuid.UUID `json:"id"`
            Email       string    `json:"email"`
            DisplayName *string   `json:"displayName,omitempty"`
            IsActive    bool      `json:"isActive"`
        }{
            ID:          a.ID().UUID(),
            Email:       a.Email().String(),
            DisplayName: a.DisplayName(),
            IsActive:    a.IsActive(),
        },
    }
}
