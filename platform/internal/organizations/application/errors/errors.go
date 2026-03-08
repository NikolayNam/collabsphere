package errors

import "github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"

var (
	ErrValidation = fault.ErrValidation
	ErrConflict   = fault.ErrConflict
	ErrInternal   = fault.ErrInternal
	ErrNotFound   = fault.ErrNotFound
)

const (
	CodeInvalidInput                   = "ORGANIZATIONS_INVALID_INPUT"
	CodeOrganizationExists             = "ORGANIZATIONS_ALREADY_EXISTS"
	CodeOrganizationNotFound           = "ORGANIZATION_NOT_FOUND"
	CodeCooperationApplicationNotFound = "COOPERATION_APPLICATION_NOT_FOUND"
	CodeInternal                       = "INTERNAL"
)

func InvalidInput(message string, opts ...fault.Opt) error {
	opts = append([]fault.Opt{fault.Code(CodeInvalidInput)}, opts...)
	return fault.Validation(message, opts...)
}

func OrganizationAlreadyExists() error {
	return fault.Conflict("Organization already exists", fault.Code(CodeOrganizationExists))
}

func OrganizationNotFound() error {
	return fault.NotFound("Organization not found", fault.Code(CodeOrganizationNotFound))
}

func CooperationApplicationNotFound() error {
	return fault.NotFound("Cooperation application not found", fault.Code(CodeCooperationApplicationNotFound))
}

func Internal(detail string, cause error) error {
	_ = detail
	return fault.Internal("Internal error", fault.Code(CodeInternal), fault.WithCause(cause))
}
