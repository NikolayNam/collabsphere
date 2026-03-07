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

    ErrNameInvalid = errors.New("name is invalid")
    ErrSlugInvalid = errors.New("slug is invalid")

    ErrInvalidOrganizationStatus = errors.New("invalid organization status")

    ErrMembershipInvalid = errors.New("membership is invalid")
)
