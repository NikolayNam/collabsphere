package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/NikolayNam/collabsphere-go/internal/users/domain"
)

type Repository interface {
	Create(ctx context.Context, u *domain.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreateUserCmd struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     string
}

func (s *Service) CreateUser(ctx context.Context, cmd CreateUserCmd) (*domain.User, error) {
	// 1) normalize/validate email
	email, err := domain.NewEmail(cmd.Email)
	if err != nil {
		return nil, ErrValidation
	}

	if len(strings.TrimSpace(cmd.Password)) == 0 {
		return nil, ErrValidation
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrInternal
	}
	ph, err := domain.NewPasswordHash(string(hash))
	if err != nil {
		return nil, ErrInternal
	}

	// 4) optional: уникальность email (если repo ожидает org)
	exists, err := s.repo.ExistsByEmail(ctx, email.String())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}
	if exists {
		return nil, ErrConflict
	}

	// 5) create domain entity
	u, err := domain.NewUser(domain.NewUserParams{
		ID:           domain.NewUserID(),
		Email:        email,
		PasswordHash: ph,
		FirstName:    cmd.FirstName,
		LastName:     cmd.LastName,
		Phone:        cmd.Phone,
		Now:          time.Now(),
	})
	if err != nil {
		return nil, ErrValidation
	}

	// 6) persist
	if err := s.repo.Create(ctx, u); err != nil {
		// если ты маппишь уникальность в repo и он возвращает ErrConflict — ок
		if errors.Is(err, ErrConflict) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}

	return u, nil
}
