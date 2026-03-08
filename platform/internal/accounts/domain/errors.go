package domain

import "errors"

var (
	ErrNowRequired       = errors.New("now is required")
	ErrTimestampsMissing = errors.New("timestamps are required")
	ErrTimestampsInvalid = errors.New("timestamps are invalid")

	ErrUserIDEmpty       = errors.New("account id is empty")
	ErrEmailEmpty        = errors.New("email is empty")
	ErrEmailInvalid      = errors.New("email is invalid")
	ErrPasswordHashEmpty = errors.New("password hash is empty")
	ErrEmailTooLong      = errors.New("email is too long")

	ErrDisplayNameInvalid = errors.New("display name is invalid")
	ErrBioInvalid         = errors.New("bio is invalid")
	ErrPhoneInvalid       = errors.New("phone is invalid")
	ErrLocaleInvalid      = errors.New("locale is invalid")
	ErrTimezoneInvalid    = errors.New("timezone is invalid")
	ErrWebsiteInvalid     = errors.New("website is invalid")

	ErrInvalidAccountStatus             = errors.New("invalid account status")
	ErrAccountBlocked                   = errors.New("account is blocked")
	ErrAccountSuspended                 = errors.New("account is suspended")
	ErrAccountStateTransitionNotAllowed = errors.New("account state is not allowed")
)
