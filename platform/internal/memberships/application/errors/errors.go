package errors

import (
	"errors"

	"github.com/NikolayNam/collabsphere/internal/memberships/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
)

var (
	ErrValidation = fault.ErrValidation
	ErrConflict   = fault.ErrConflict
	ErrInternal   = fault.ErrInternal
	ErrNotFound   = fault.ErrNotFound
)

const (
	CodeInvalidInput          = "ORGANIZATIONS_INVALID_INPUT"
	CodeOrganizationExists    = "ORGANIZATIONS_ALREADY_EXISTS"
	CodeOrganizationNotFound  = "ORGANIZATION_NOT_FOUND"
	CodeMemberExists          = "MEMBER_EXIST"
	CodeInvitationExists      = "ORGANIZATION_INVITATION_EXISTS"
	CodeInvitationNotFound    = "ORGANIZATION_INVITATION_NOT_FOUND"
	CodeInvitationExpired     = "ORGANIZATION_INVITATION_EXPIRED"
	CodeInvitationEmail       = "ORGANIZATION_INVITATION_EMAIL_MISMATCH"
	CodeAccessRequestExists   = "ORGANIZATION_ACCESS_REQUEST_EXISTS"
	CodeAccessRequestNotFound = "ORGANIZATION_ACCESS_REQUEST_NOT_FOUND"
	CodeAccessRequestClosed   = "ORGANIZATION_ACCESS_REQUEST_ALREADY_REVIEWED"
	CodeInternal              = "INTERNAL"
)

func InvalidInput(message string, opts ...fault.Opt) error {
	opts = append([]fault.Opt{fault.Code(CodeInvalidInput)}, opts...)
	return fault.Validation(message, opts...)
}

func MemberAlreadyExists() error {
	return fault.Conflict("Member already exists", fault.Code(CodeMemberExists))
}

func InvitationAlreadyExists() error {
	return fault.Conflict("Organization invitation already exists", fault.Code(CodeInvitationExists))
}

func InvitationNotFound() error {
	return fault.NotFound("Organization invitation not found", fault.Code(CodeInvitationNotFound))
}

func InvitationExpired() error {
	return fault.Conflict("Organization invitation is expired", fault.Code(CodeInvitationExpired))
}

func InvitationEmailMismatch() error {
	return fault.Forbidden("Invitation email does not match authenticated account", fault.Code(CodeInvitationEmail))
}

func AccessRequestAlreadyExists() error {
	return fault.Conflict("Organization access request already exists", fault.Code(CodeAccessRequestExists))
}

func AccessRequestNotFound() error {
	return fault.NotFound("Organization access request not found", fault.Code(CodeAccessRequestNotFound))
}

func AccessRequestAlreadyReviewed() error {
	return fault.Conflict("Organization access request is already reviewed", fault.Code(CodeAccessRequestClosed))
}

func OrganizationAlreadyExists() error {
	return fault.Conflict("Organization already exists", fault.Code(CodeOrganizationExists))
}

func OrganizationNotFound() error {
	return fault.NotFound("Organization not found", fault.Code(CodeOrganizationNotFound))
}

func Internal(detail string, cause error) error {
	_ = detail
	return fault.Internal("Internal error", fault.Code(CodeInternal), fault.WithCause(cause))
}

func FromDomain(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, domain.ErrOrganizationRoleCodeInvalid), errors.Is(err, domain.ErrOrganizationRoleCodeReserved):
		return InvalidInput(err.Error())
	case errors.Is(err, domain.ErrOrganizationRoleNameInvalid), errors.Is(err, domain.ErrOrganizationRoleBaseRoleInvalid):
		return InvalidInput(err.Error())
	case errors.Is(err, domain.ErrOrganizationRoleInvalid):
		return InvalidInput(err.Error())
	default:
		return err
	}
}
