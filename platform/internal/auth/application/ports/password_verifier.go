package ports

import accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"

type PasswordVerifier interface {
	Verify(hash accdomain.PasswordHash, raw string) error
}
