package access

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	accesspolicy "github.com/NikolayNam/collabsphere/internal/iam/access"
	memberdomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

func RequireOrganizationAccess(ctx context.Context, organizations ports.OrganizationReader, memberships ports.MembershipReader, organizationID orgdomain.OrganizationID, actorID accdomain.AccountID, requireManage bool) error {
	_, member, err := loadOrganizationMember(ctx, organizations, memberships, organizationID, actorID)
	if err != nil {
		return err
	}
	if requireManage && !accesspolicy.HasOrganizationPermission(member.Role(), accesspolicy.PermissionOrganizationManageCatalog) {
		return catalogerrors.AccessDenied()
	}
	return nil
}

func RequireOrganizationEmployeeAccess(ctx context.Context, organizations ports.OrganizationReader, memberships ports.MembershipReader, organizationID orgdomain.OrganizationID, actorID accdomain.AccountID) error {
	_, member, err := loadOrganizationMember(ctx, organizations, memberships, organizationID, actorID)
	if err != nil {
		return err
	}
	if !accesspolicy.HasOrganizationPermission(member.Role(), accesspolicy.PermissionOrganizationEmployeeAccess) {
		return catalogerrors.AccessDenied()
	}
	return nil
}

func loadOrganizationMember(ctx context.Context, organizations ports.OrganizationReader, memberships ports.MembershipReader, organizationID orgdomain.OrganizationID, actorID accdomain.AccountID) (any, *memberdomain.Membership, error) {
	if organizationID.IsZero() {
		return nil, nil, catalogerrors.InvalidInput("Organization is required")
	}
	if actorID.IsZero() {
		return nil, nil, catalogerrors.InvalidInput("Actor account is required")
	}

	organization, err := organizations.GetByID(ctx, organizationID)
	if err != nil {
		return nil, nil, err
	}
	if organization == nil {
		return nil, nil, catalogerrors.OrganizationNotFound()
	}

	member, err := memberships.GetMemberByAccount(ctx, organizationID, actorID)
	if err != nil {
		return nil, nil, catalogerrors.Internal("get organization membership", err)
	}
	if member == nil || !member.IsActive() || member.IsRemoved() {
		return nil, nil, catalogerrors.AccessDenied()
	}
	return organization, member, nil
}
