package errors

import "github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"

var (
	ErrValidation = fault.ErrValidation
	ErrConflict   = fault.ErrConflict
	ErrInternal   = fault.ErrInternal
	ErrNotFound   = fault.ErrNotFound
	ErrForbidden  = fault.ErrForbidden
)

const (
	CodeInvalidInput      = "GROUPS_INVALID_INPUT"
	CodeGroupExists       = "GROUPS_ALREADY_EXISTS"
	CodeGroupNotFound     = "GROUP_NOT_FOUND"
	CodeGroupMemberExists = "GROUP_MEMBER_ALREADY_EXISTS"
	CodeAccessDenied      = "GROUP_ACCESS_DENIED"
	CodeInternal          = "INTERNAL"
)

func InvalidInput(message string, opts ...fault.Opt) error {
	opts = append([]fault.Opt{fault.Code(CodeInvalidInput)}, opts...)
	return fault.Validation(message, opts...)
}

func GroupAlreadyExists() error {
	return fault.Conflict("Group already exists", fault.Code(CodeGroupExists))
}

func GroupNotFound() error {
	return fault.NotFound("Group not found", fault.Code(CodeGroupNotFound))
}

func GroupMemberAlreadyExists() error {
	return fault.Conflict("Group member already exists", fault.Code(CodeGroupMemberExists))
}

func AccessDenied() error {
	return fault.Forbidden("Access denied", fault.Code(CodeAccessDenied))
}

func Internal(detail string, cause error) error {
	_ = detail
	return fault.Internal("Internal error", fault.Code(CodeInternal), fault.WithCause(cause))
}
