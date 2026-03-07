package domain

import "errors"

var (
	ErrNowRequired             = errors.New("now is required")
	ErrSessionIDRequired       = errors.New("session id is required")
	ErrAccountIDRequired       = errors.New("account id is required")
	ErrTokenHashRequired       = errors.New("token hash is required")
	ErrSessionExpiresAtInvalid = errors.New("session expires_at is invalid")
	ErrRefreshTokenInvalid     = errors.New("refresh token is invalid")
)
