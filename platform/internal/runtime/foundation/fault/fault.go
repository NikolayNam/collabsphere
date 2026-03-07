package fault

import (
	"errors"
	"fmt"
)

// Kind — классификация ошибок приложения (не инфраструктуры).
// На базе Kind мы маппим ошибку в HTTP status (в translator для Huma).
type Kind uint8

const (
	KindUnknown Kind = iota
	KindValidation
	KindConflict
	KindNotFound
	KindUnauthorized
	KindForbidden
	KindTooManyRequests
	KindUnavailable
	KindInternal
)

func (k Kind) String() string {
	switch k {
	case KindValidation:
		return "validation"
	case KindConflict:
		return "conflict"
	case KindNotFound:
		return "not_found"
	case KindUnauthorized:
		return "unauthorized"
	case KindForbidden:
		return "forbidden"
	case KindTooManyRequests:
		return "too_many_requests"
	case KindUnavailable:
		return "unavailable"
	case KindInternal:
		return "internal"
	default:
		return "unknown"
	}
}

// FieldError — детализация ошибок для конкретных полей (валидация и т.п.).
type FieldError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// Error — типизированная ошибка приложения.
// Важно: Message должен быть безопасным для отдачи клиенту.
// Cause — внутренняя причина (DB/IO/etc), не для клиента.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Fields  []FieldError
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}

	// Делаем строку безопасной: не включаем Cause, не раскрываем детали.
	switch {
	case e.Code != "" && e.Message != "":
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	case e.Message != "":
		return e.Message
	case e.Code != "":
		return e.Code
	default:
		return e.Kind.String()
	}
}

func (e *Error) Unwrap() error { return e.Cause }

// Is делает errors.Is полезным для сравнений по Kind и/или Code.
// Пример: errors.Is(err, apperr.ErrValidation) == true
// Пример: errors.Is(err, apperr.New(KindValidation, "", apperr.Code("x"))) можно проверить по Code.
func (e *Error) Is(target error) bool {
	var t *Error
	ok := errors.As(target, &t)
	if !ok || e == nil || t == nil {
		return false
	}

	// Если в target.Kind задан (не Unknown), то Kind должен совпасть.
	if t.Kind != KindUnknown && e.Kind != t.Kind {
		return false
	}

	// Если в target.Code задан, то Code должен совпасть.
	if t.Code != "" && e.Code != t.Code {
		return false
	}

	return true
}

// Opt — функциональные опции для Error.
type Opt func(*Error)

func Code(code string) Opt {
	return func(e *Error) { e.Code = code }
}

func WithCause(err error) Opt {
	return func(e *Error) { e.Cause = err }
}

func Field(field, message string) Opt {
	return func(e *Error) {
		e.Fields = append(e.Fields, FieldError{Field: field, Message: message})
	}
}

func Fields(items ...FieldError) Opt {
	return func(e *Error) { e.Fields = append(e.Fields, items...) }
}

func New(kind Kind, msg string, opts ...Opt) *Error {
	e := &Error{
		Kind:    kind,
		Message: msg,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(e)
		}
	}
	return e
}

// Удобные конструкторы.
func Validation(msg string, opts ...Opt) *Error      { return New(KindValidation, msg, opts...) }
func Conflict(msg string, opts ...Opt) *Error        { return New(KindConflict, msg, opts...) }
func NotFound(msg string, opts ...Opt) *Error        { return New(KindNotFound, msg, opts...) }
func Unauthorized(msg string, opts ...Opt) *Error    { return New(KindUnauthorized, msg, opts...) }
func Forbidden(msg string, opts ...Opt) *Error       { return New(KindForbidden, msg, opts...) }
func TooManyRequests(msg string, opts ...Opt) *Error { return New(KindTooManyRequests, msg, opts...) }
func Unavailable(msg string, opts ...Opt) *Error     { return New(KindUnavailable, msg, opts...) }
func Internal(msg string, opts ...Opt) *Error        { return New(KindInternal, msg, opts...) }

// Сентинелы по Kind: для errors.Is(err, apperr.ErrValidation) и т.п.
// Эти значения НЕ предназначены для прямой отдачи клиенту (нет Message).
var (
	ErrValidation      = &Error{Kind: KindValidation}
	ErrConflict        = &Error{Kind: KindConflict}
	ErrNotFound        = &Error{Kind: KindNotFound}
	ErrUnauthorized    = &Error{Kind: KindUnauthorized}
	ErrForbidden       = &Error{Kind: KindForbidden}
	ErrTooManyRequests = &Error{Kind: KindTooManyRequests}
	ErrUnavailable     = &Error{Kind: KindUnavailable}
	ErrInternal        = &Error{Kind: KindInternal}
)

// As вытаскивает *apperr.Error из цепочки.
func As(err error) (*Error, bool) {
	if e, ok := errors.AsType[*Error](err); ok && e != nil {
		return e, true
	}
	return nil, false
}
