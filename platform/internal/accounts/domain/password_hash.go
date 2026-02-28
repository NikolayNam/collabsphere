package domain

type PasswordHash string

func NewPasswordHash(hash string) (PasswordHash, error) {
	if hash == "" {
		return "", ErrPasswordHashEmpty
	}
	return PasswordHash(hash), nil
}

func (h PasswordHash) String() string { return string(h) }
func (h PasswordHash) IsZero() bool   { return string(h) == "" }
