package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres/mapper"
)

func (r *AccountRepo) GetByID(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	var row mapper.AccountRow

	err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.accounts AS a").
		Select("a.id, a.email, zi.external_subject AS zitadel_user_id, a.display_name, a.avatar_object_id, a.bio, a.phone, a.locale, a.timezone, a.website, a.is_active, a.created_at, a.updated_at, pc.password_hash").
		Joins("LEFT JOIN auth.password_credentials AS pc ON pc.account_id = a.id").
		Joins("LEFT JOIN LATERAL (SELECT ei.external_subject FROM auth.external_identities AS ei WHERE ei.account_id = a.id AND ei.provider = 'zitadel' ORDER BY ei.last_login_at DESC NULLS LAST, ei.created_at DESC LIMIT 1) AS zi ON true").
		Where("a.id = ?", id.UUID()).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mapper.ToDomainAccount(&row)
}

func (r *AccountRepo) GetByEmail(ctx context.Context, email domain.Email) (*domain.Account, error) {
	var row mapper.AccountRow

	err := r.dbFrom(ctx).WithContext(ctx).
		Table("iam.accounts AS a").
		Select("a.id, a.email, zi.external_subject AS zitadel_user_id, a.display_name, a.avatar_object_id, a.bio, a.phone, a.locale, a.timezone, a.website, a.is_active, a.created_at, a.updated_at, pc.password_hash").
		Joins("LEFT JOIN auth.password_credentials AS pc ON pc.account_id = a.id").
		Joins("LEFT JOIN LATERAL (SELECT ei.external_subject FROM auth.external_identities AS ei WHERE ei.account_id = a.id AND ei.provider = 'zitadel' ORDER BY ei.last_login_at DESC NULLS LAST, ei.created_at DESC LIMIT 1) AS zi ON true").
		Where("a.email = ?", email.String()).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mapper.ToDomainAccount(&row)
}
