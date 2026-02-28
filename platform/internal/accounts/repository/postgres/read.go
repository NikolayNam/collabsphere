package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres/mapper"
)

func (r *AccountRepo) GetByID(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	var m dbmodel.Account

	err := r.db.WithContext(ctx).
		Take(&m, "id = ?", id.UUID()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mapper.ToDomainAccount(&m)
}
func (r *AccountRepo) GetByEmail(ctx context.Context, email domain.Email) (*domain.Account, error) {
	var m dbmodel.Account
	err := r.db.WithContext(ctx).
		Where("email = ?", email.String()).
		Take(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToDomainAccount(&m)
}

func (r *AccountRepo) ExistsByEmail(ctx context.Context, email domain.Email) (bool, error) {
	var one int

	err := r.db.WithContext(ctx).
		Model(&dbmodel.Account{}).
		Select("1").
		Where("email = ?", email.String()).
		Limit(1).
		Take(&one).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
