package dto

import "github.com/google/uuid"

type AccountResponse struct {
    Status int `json:"-"`
    Body   struct {
        ID          uuid.UUID `json:"id"`
        Email       string    `json:"email"`
        DisplayName *string   `json:"displayName,omitempty"`
        IsActive    bool      `json:"isActive"`
    }
}
