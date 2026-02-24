package postgres

import (
	"context"

	"github.com/NikolayNam/collabsphere-go/internal/users/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere-go/internal/users/repository/postgres/mapper"
	"gorm.io/gorm"

	"github.com/NikolayNam/collabsphere-go/internal/users/domain"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) ByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	var m dbmodel.User
	err := r.db.WithContext(ctx).First(&m, "id = ?", id.UUID()).Error
	if err != nil {
		return nil, err
	}
	return mapper.ToDomainUser(&m)
}

func (r *UserRepo) ByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	var m dbmodel.User
	err := r.db.WithContext(ctx).First(&m, "email = ?", email.String()).Error
	if err != nil {
		return nil, err
	}
	return mapper.ToDomainUser(&m)
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	m, err := mapper.ToDBUser(u)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	// ВАЖНО: это корректно ТОЛЬКО если в таблице users реально есть organization_id.
	// Если у тебя роли/пользователи не привязаны к org, то интерфейс сервиса неправильный.
	var count int64
	err := r.db.WithContext(ctx).
		Model(&dbmodel.User{}).
		Where("email = ?", email).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
