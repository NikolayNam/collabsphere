package domain

import "errors"

var (
	ErrGroupIDEmpty            = errors.New("group id is empty")
	ErrGroupNameInvalid        = errors.New("group name is invalid")
	ErrGroupSlugInvalid        = errors.New("group slug is invalid")
	ErrGroupDescriptionInvalid = errors.New("group description is invalid")
	ErrGroupRoleInvalid        = errors.New("group role is invalid")
	ErrNowRequired             = errors.New("now is required")
	ErrTimestampsMissing       = errors.New("timestamps are required")
	ErrTimestampsInvalid       = errors.New("timestamps are invalid")
	ErrAccountMemberInvalid    = errors.New("group account member is invalid")
	ErrOrgMemberInvalid        = errors.New("group organization member is invalid")
)
