package bcrypt

import (
	"golang.org/x/crypto/bcrypt"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
)

func (h *Hasher) Verify(hash accdomain.PasswordHash, raw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash.String()), []byte(raw))
}
