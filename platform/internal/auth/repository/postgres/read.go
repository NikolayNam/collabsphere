package postgres

import (
	"context"
	"errors"

	"github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/NikolayNam/collabsphere/internal/auth/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/auth/repository/postgres/mapper"
	"gorm.io/gorm"
)

func (r *SessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshSession, error) {
	var model dbmodel.RefreshSession

	err := r.dbFrom(ctx).WithContext(ctx).
		Table("auth.refresh_sessions AS s").
		Select("s.*").
		Joins("JOIN auth.refresh_session_tokens AS t ON t.session_id = s.id").
		Where("t.token_hash = ?", tokenHash).
		Take(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mapper.ToDomainRefreshSession(&model)
}
