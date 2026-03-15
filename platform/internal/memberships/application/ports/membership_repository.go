package ports

import (
	"context"
	"time"

	accDomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type MembershipRepository interface {
	AddMember(ctx context.Context, orgID orgDomain.OrganizationID, m *memberDomain.Membership) error
	SaveMember(ctx context.Context, orgID orgDomain.OrganizationID, m *memberDomain.Membership) error
	GetMemberByAccount(ctx context.Context, orgID orgDomain.OrganizationID, accountID accDomain.AccountID) (*memberDomain.Membership, error)
	GetMemberByID(ctx context.Context, orgID orgDomain.OrganizationID, membershipID uuid.UUID) (*memberDomain.Membership, error)
	CountActiveMembersByRole(ctx context.Context, orgID orgDomain.OrganizationID, role memberDomain.MembershipRole) (int64, error)
	ListMembers(ctx context.Context, orgID orgDomain.OrganizationID) ([]memberDomain.MemberView, error)
	CreateInvitation(ctx context.Context, invitation *memberDomain.OrganizationInvitation) error
	SaveInvitation(ctx context.Context, invitation *memberDomain.OrganizationInvitation) error
	GetInvitationByTokenHash(ctx context.Context, tokenHash string) (*memberDomain.OrganizationInvitation, error)
	ListInvitations(ctx context.Context, orgID orgDomain.OrganizationID) ([]memberDomain.OrganizationInvitation, error)
	RevokeExpiredPendingInvitations(ctx context.Context, orgID orgDomain.OrganizationID, email accDomain.Email, actorAccountID uuid.UUID, now time.Time) error
	CreateAccessRequest(ctx context.Context, req *memberDomain.OrganizationAccessRequest) error
	SaveAccessRequest(ctx context.Context, req *memberDomain.OrganizationAccessRequest) error
	GetAccessRequestByID(ctx context.Context, orgID orgDomain.OrganizationID, requestID uuid.UUID) (*memberDomain.OrganizationAccessRequest, error)
	ListAccessRequests(ctx context.Context, orgID orgDomain.OrganizationID) ([]memberDomain.OrganizationAccessRequest, error)
}
