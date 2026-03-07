package access

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	memberdomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

func RequireOrganizationAccess(ctx context.Context, organizations ports.OrganizationReader, memberships ports.MembershipReader, organizationID orgdomain.OrganizationID, actorID accdomain.AccountID, requireOwner bool) error {
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

	members, err := memberships.ListMembers(ctx, organizationID)
	if err != nil {
		return catalogerrors.Internal("list organization members", err)
	}

	for _, member := range members {
		if member.AccountID != actorID.UUID() || !member.IsActive {
			continue
		}
		if requireOwner && member.Role != string(memberdomain.MembershipRoleOwner) {
			return catalogerrors.AccessDenied()
		}
		return nil
	}

	return catalogerrors.AccessDenied()
}
