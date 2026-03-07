package domain

import "strings"

type RefreshToken string

func NewRefreshToken(raw string) (RefreshToken, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", ErrRefreshTokenInvalid
	}
	return RefreshToken(s), nil
}

func (t RefreshToken) String() string {
	return string(t)
}

func (t RefreshToken) IsZero() bool {
	return strings.TrimSpace(string(t)) == ""
}
