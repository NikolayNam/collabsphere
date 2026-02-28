package errors

import "github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"

var (
	ErrValidation = fault.ErrValidation
	ErrConflict   = fault.ErrConflict
	ErrInternal   = fault.ErrInternal
	ErrNotFound   = fault.ErrNotFound
)

const (
	CodeInvalidInput    = "ACCOUNTS_INVALID_INPUT"
	CodeAccountExists   = "ACCOUNTS_ALREADY_EXISTS"
	CodeAccountNotFound = "ACCOUNT_NOT_FOUND"
	CodeInternal        = "INTERNAL"
)

func InvalidInput(message string, opts ...fault.Opt) error {
	// гарантируем, что code выставлен
	opts = append([]fault.Opt{fault.Code(CodeInvalidInput)}, opts...)
	return fault.Validation(message, opts...)
}

// AccountAlreadyExists — ошибка 409.
func AccountAlreadyExists() error {
	return fault.Conflict("Account already exists", fault.Code(CodeAccountExists))
}

// AccountNotFound — ошибка 404.
func AccountNotFound() error {
	return fault.NotFound("Account not found", fault.Code(CodeAccountNotFound))
}

// Internal — ошибка 500. detail/cause нужны для диагностики, но клиенту это показывать нельзя.
// Важно: message здесь безопасный, подробности оставляем в cause.
func Internal(detail string, cause error) error {
	_ = detail // detail можно использовать позже в логировании/метриках (или включать в cause через fmt.Errorf)
	return fault.Internal("Internal error", fault.Code(CodeInternal), fault.WithCause(cause))
}
