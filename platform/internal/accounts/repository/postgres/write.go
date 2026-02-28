package postgres

import (
	"context"
	"errors"

	apperrors "github.com/NikolayNam/collabsphere/internal/accounts/application/errors"

	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres/mapper"
)

func (r *AccountRepo) Create(ctx context.Context, account *domain.Account) error {
	if account == nil {
		return errors.New("account is nil")
	}

	m := mapper.ToDBAccountForCreate(account)
	if m == nil {
		return errors.New("db account model is nil")
	}

	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.AccountAlreadyExists()
		}
		return err
	}
	return nil

}
