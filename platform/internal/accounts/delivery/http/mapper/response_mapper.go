package mapper

import (
	"github.com/google/uuid"

	"github.com/NikolayNam/collabsphere/internal/accounts/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
)

func ToAccountResponse(a *domain.Account) *dto.AccountResponse {
	if a == nil {
		return nil
	}

	return &dto.AccountResponse{
		Body: struct {
			ID        uuid.UUID `json:"id"`
			Email     string    `json:"email"`
			FirstName string    `json:"first_name"`
			LastName  string    `json:"last_name"`
			Status    string    `json:"status"`
		}{
			ID:        a.ID().UUID(),
			Email:     a.Email().String(),
			FirstName: a.FirstName(),
			LastName:  a.LastName(),
			Status:    string(a.Status()),
		},
	}
}
