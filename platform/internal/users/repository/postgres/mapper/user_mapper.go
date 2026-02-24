package mapper

import (
	"strings"
	"time"

	"github.com/NikolayNam/collabsphere-go/internal/users/domain"
	basemodel "github.com/NikolayNam/collabsphere-go/internal/users/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere-go/shared/contracts/persistence/dbmodel"
)

func ToDomainUser(m *basemodel.User) (*domain.User, error) {
	if m == nil {
		return nil, nil
	}

	id, err := domain.UserIDFromUUID(m.ID)
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

	return domain.RehydrateUser(domain.RehydrateUserParams{
		id,
		email,
		hash,
		m.FirstName,
		m.LastName,
		ptrToString(m.Phone), // *string -> string
		m.IsActive,
		m.CreatedAt,
		m.UpdatedAt,
	})
}

func ToDBUser(u *domain.User) (*basemodel.User, error) {
	if u == nil {
		return nil, nil
	}

	now := time.Now()

	return &basemodel.User{
		BaseModel: dbmodel.BaseModel{
			UUIDPK: dbmodel.UUIDPK{
				ID: u.ID().UUID(),
			},
			Timestamps: dbmodel.Timestamps{
				CreatedAt: nonZeroOrNow(u.CreatedAt(), now),
				UpdatedAt: nonZeroOrNow(u.UpdatedAt(), now),
			},
			// Blame: оставляем persistence callbacks
		},
		Email:        u.Email().String(),
		PasswordHash: u.PasswordHash().String(),
		FirstName:    u.FirstName(),
		LastName:     u.LastName(),
		Phone:        stringToNilPtr(u.Phone()),
		IsActive:     u.IsActive(),
	}, nil
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func stringToNilPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return new(s)
}

func nonZeroOrNow(t time.Time, now time.Time) time.Time {
	if t.IsZero() {
		return now
	}
	return t
}
