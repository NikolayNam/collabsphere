package mapper

import (
	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres/dbmodel"
)

func ToDomainAccount(m *dbmodel.Account) (*domain.Account, error) {
	if m == nil {
		return nil, nil
	}

	id, err := domain.AccountIDFromUUID(m.ID)
	if err != nil {
		return nil, err
	}

	email, err := domain.NewEmail(m.Email)
	if err != nil {
		return nil, err
	}

	hash, err := domain.NewPasswordHash(m.PasswordHash)
	if err != nil {
		return nil, err
	}

	status, err := domain.NewAccountStatus(m.Status)
	if err != nil {
		return nil, err
	}

	return domain.RehydrateAccount(domain.RehydrateAccountParams{
		ID:           id,
		Email:        email,
		PasswordHash: hash,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		Status:       status,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	})
}
