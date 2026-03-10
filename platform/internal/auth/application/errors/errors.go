package errors

import "github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"

var (
	ErrValidation   = fault.ErrValidation
	ErrConflict     = fault.ErrConflict
	ErrInternal     = fault.ErrInternal
	ErrNotFound     = fault.ErrNotFound
	ErrUnauthorized = fault.ErrUnauthorized
	ErrForbidden    = fault.ErrForbidden
	ErrUnavailable  = fault.ErrUnavailable
)

const (
	CodeInvalidInput    = "AUTH_INVALID_INPUT"
	CodeUnauthorized    = "AUTH_UNAUTHORIZED"
	CodeForbidden       = "AUTH_FORBIDDEN"
	CodeRefreshInvalid  = "AUTH_REFRESH_INVALID"
	CodeSessionNotFound = "AUTH_SESSION_NOT_FOUND"
	CodeOIDCUnavailable = "AUTH_OIDC_UNAVAILABLE"
	CodeInternal        = "INTERNAL"
)

func InvalidInput(message string, opts ...fault.Opt) error {
	opts = append([]fault.Opt{fault.Code(CodeInvalidInput)}, opts...)
	return fault.Validation(message, opts...)
}

func Unauthorized(message string, opts ...fault.Opt) error {
	opts = append([]fault.Opt{fault.Code(CodeUnauthorized)}, opts...)
	return fault.Unauthorized(message, opts...)
}

func Forbidden(message string, opts ...fault.Opt) error {
	opts = append([]fault.Opt{fault.Code(CodeForbidden)}, opts...)
	return fault.Forbidden(message, opts...)
}

func RefreshTokenInvalid() error {
	return fault.Unauthorized("Invalid refresh token", fault.Code(CodeRefreshInvalid))
}

func SessionNotFound() error {
	return fault.NotFound("Session not found", fault.Code(CodeSessionNotFound))
}

func Unavailable(message string, opts ...fault.Opt) error {
	opts = append([]fault.Opt{fault.Code(CodeOIDCUnavailable)}, opts...)
	return fault.Unavailable(message, opts...)
}

func Internal(detail string, cause error) error {
	_ = detail
	return fault.Internal("Internal error", fault.Code(CodeInternal), fault.WithCause(cause))
}
