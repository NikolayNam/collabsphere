package ports

import "github.com/NikolayNam/collabsphere/internal/accounts/domain"

type PasswordHasher interface {
	Hash(raw string) (domain.PasswordHash, error)
}
