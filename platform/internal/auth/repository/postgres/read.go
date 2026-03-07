package postgres

import (
	"context"
	"errors"

	"github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/NikolayNam/collabsphere/internal/auth/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/auth/repository/postgres/mapper"
	"gorm.io/gorm"
)

func (r *SessionRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshSession, error) {
	var m dbmodel.RefreshSession

	err := r.dbFrom(ctx).WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		Take(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mapper.ToDomainRefreshSession(&m)
}
