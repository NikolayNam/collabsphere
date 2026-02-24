package domain

import "errors"

// PasswordHash — доменный тип (не “просто строка”).
// В домене мы не обязаны знать алгоритм (bcrypt/argon2) — это outside world.
type PasswordHash string

func NewPasswordHash(hash string) (PasswordHash, error) {
	if hash == "" {
		return "", errors.New("password hash is empty")
	}
	return PasswordHash(hash), nil
}

func (h PasswordHash) String() string { return string(h) }
func (h PasswordHash) IsZero() bool   { return string(h) == "" }
