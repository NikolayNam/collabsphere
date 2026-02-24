package domain

import "errors"

const (
	CodeInvalidEmail         ErrorCode = "INVALID_EMAIL"
	CodeInvalidLegalName     ErrorCode = "INVALID_LEGAL_NAME"
	CodeAlreadyVerified      ErrorCode = "ALREADY_VERIFIED"
	CodeAlreadyDeactivated   ErrorCode = "ALREADY_DEACTIVATED"
	CodeCreatorLoginRequired ErrorCode = "CREATOR_LOGIN_REQUIRED"
	CodeDomain               ErrorCode = "DOMAIN_ERROR"
)

type (
	// ErrorCode — стабильные коды доменных ошибок.
	ErrorCode string
)

// Error — единый тип доменной ошибки.
type Error struct {
	Code    ErrorCode
	Message string
}

func (e Error) Error() string {
	return e.Message
}

// Is позволяет использовать errors.Is(...)
func (e Error) Is(target error) bool {
	var t Error
	ok := errors.As(target, &t)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func NewInvalidEmail() error {
	return Error{
		Code:    CodeInvalidEmail,
		Message: "Неверный формат email",
	}
}

func NewInvalidLegalName() error {
	return Error{
		Code:    CodeInvalidLegalName,
		Message: "Некорректное юридическое наименование",
	}
}

func NewAlreadyVerified() error {
	return Error{
		Code:    CodeAlreadyVerified,
		Message: "Компания уже верифицирована",
	}
}

func NewAlreadyDeactivated() error {
	return Error{
		Code:    CodeAlreadyDeactivated,
		Message: "Компания деактивирована",
	}
}

func NewCreatorLoginRequired() error {
	return Error{
		Code:    CodeCreatorLoginRequired,
		Message: "Логин создателя обязателен",
	}
}

func New(msg string) error {
	return Error{
		Code:    CodeDomain,
		Message: msg,
	}
}
