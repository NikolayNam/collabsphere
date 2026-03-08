package mapper

import (
	"time"

	"github.com/google/uuid"

	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres/dbmodel"
)

func ToDBAccountForCreate(a *domain.Account) *dbmodel.Account {
	if a == nil {
		return nil
	}

	updatedAt := a.CreatedAt()
	if a.UpdatedAt() != nil {
		updatedAt = *a.UpdatedAt()
	}

	return &dbmodel.Account{
		ID:             a.ID().UUID(),
		Email:          a.Email().String(),
		DisplayName:    a.DisplayName(),
		AvatarObjectID: a.AvatarObjectID(),
		Bio:            a.Bio(),
		Phone:          a.Phone(),
		Locale:         a.Locale(),
		Timezone:       a.Timezone(),
		Website:        a.Website(),
		IsActive:       a.IsActive(),
		CreatedAt:      a.CreatedAt(),
		UpdatedAt:      updatedAt,
	}
}

func ToDBPasswordCredentialForCreate(a *domain.Account) *dbmodel.PasswordCredential {
	if a == nil {
		return nil
	}

	updatedAt := a.CreatedAt()
	if a.UpdatedAt() != nil {
		updatedAt = *a.UpdatedAt()
	}

	return &dbmodel.PasswordCredential{
		AccountID:    a.ID().UUID(),
		PasswordHash: a.PasswordHash().String(),
		CreatedAt:    a.CreatedAt(),
		UpdatedAt:    updatedAt,
	}
}

type AccountRow struct {
	ID             uuid.UUID
	Email          string
	DisplayName    *string
	AvatarObjectID *uuid.UUID
	Bio            *string
	Phone          *string
	Locale         *string
	Timezone       *string
	Website        *string
	PasswordHash   string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func ToDomainAccount(row *AccountRow) (*domain.Account, error) {
	if row == nil {
		return nil, nil
	}

	id, err := domain.AccountIDFromUUID(row.ID)
	if err != nil {
		return nil, err
	}

	email, err := domain.NewEmail(row.Email)
	if err != nil {
		return nil, err
	}

	hash, err := domain.NewPasswordHash(row.PasswordHash)
	if err != nil {
		return nil, err
	}

	return domain.RehydrateAccount(domain.RehydrateAccountParams{
		ID:             id,
		Email:          email,
		PasswordHash:   hash,
		DisplayName:    row.DisplayName,
		AvatarObjectID: row.AvatarObjectID,
		Bio:            row.Bio,
		Phone:          row.Phone,
		Locale:         row.Locale,
		Timezone:       row.Timezone,
		Website:        row.Website,
		IsActive:       row.IsActive,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	})
}
