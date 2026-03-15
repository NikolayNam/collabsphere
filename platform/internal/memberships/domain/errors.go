package domain

import "errors"

var (
	ErrNowRequired       = errors.New("now is required")
	ErrTimestampsMissing = errors.New("timestamps are required")
	ErrTimestampsInvalid = errors.New("timestamps are invalid")

	ErrOrganizationIDEmpty = errors.New("organization id is empty")

	ErrEmailEmpty   = errors.New("email is empty")
	ErrEmailInvalid = errors.New("email is invalid")
	ErrEmailTooLong = errors.New("email is too long")

	ErrLegalNameInvalid   = errors.New("legal name is invalid")
	ErrDisplayNameInvalid = errors.New("display name is invalid")

	ErrInvalidOrganizationStatus = errors.New("invalid organization status")

	ErrMembershipInvalid = errors.New("membership is invalid")

	ErrOrganizationRoleInvalid         = errors.New("organization role is invalid")
	ErrOrganizationRoleCodeInvalid    = errors.New("organization role code is invalid")
	ErrOrganizationRoleCodeReserved   = errors.New("organization role code is reserved for system roles")
	ErrOrganizationRoleNameInvalid    = errors.New("organization role name is invalid")
	ErrOrganizationRoleBaseRoleInvalid = errors.New("organization role base role is invalid")
	ErrOrganizationRoleNotFound       = errors.New("organization role not found")
	ErrOrganizationRoleInUse          = errors.New("organization role is in use and cannot be deleted")
)
