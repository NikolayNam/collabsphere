package access

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

func RequireOrganizationAccess(ctx context.Context, organizations ports.OrganizationReader, memberships ports.MembershipReader, organizationID orgdomain.OrganizationID, actorID accdomain.AccountID, requireManage bool) error {
	if organizationID.IsZero() {
		return catalogerrors.InvalidInput("Organization is required")
	}
	if actorID.IsZero() {
		return catalogerrors.InvalidInput("Actor account is required")
	}

	organization, err := organizations.GetByID(ctx, organizationID)
	if err != nil {
		return err
	}
	if organization == nil {
		return catalogerrors.OrganizationNotFound()
	}

	member, err := memberships.GetMemberByAccount(ctx, organizationID, actorID)
	if err != nil {
		return catalogerrors.Internal("get organization membership", err)
	}
	if member == nil || !member.IsActive() || member.IsRemoved() {
		return catalogerrors.AccessDenied()
	}
	if requireManage && !member.Role().CanManageCatalog() {
		return catalogerrors.AccessDenied()
	}
	return nil
}
