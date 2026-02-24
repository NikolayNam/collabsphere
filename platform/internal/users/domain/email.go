package domain

import (
	"errors"
	"regexp"
	"strings"
)

type Email string

func NewEmail(raw string) (Email, error) {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return "", errors.New("email is empty")
	}

	// Практичная проверка (не RFC-идеал, но защищает от мусора).
	re := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	if !re.MatchString(s) {
		return "", errors.New("email is invalid")
	}

	return Email(s), nil
}

func (e Email) String() string { return string(e) }
func (e Email) IsZero() bool   { return strings.TrimSpace(string(e)) == "" }
