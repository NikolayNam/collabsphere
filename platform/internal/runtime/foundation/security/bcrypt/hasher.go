package bcrypt

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/NikolayNam/collabsphere/internal/accounts/domain"
)

type Hasher struct {
	cost int
}

func NewBcryptHasher() *Hasher {
	return &Hasher{cost: bcrypt.DefaultCost}
}

func (h *Hasher) Hash(raw string) (domain.PasswordHash, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("password is empty")
	}

	b, err := bcrypt.GenerateFromPassword([]byte(raw), h.cost)
	if err != nil {
		return "", err
	}

	return domain.NewPasswordHash(string(b))
}
