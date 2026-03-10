package application

import (
	"context"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	memberports "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	uploadports "github.com/NikolayNam/collabsphere/internal/uploads/application/ports"
	uploaddomain "github.com/NikolayNam/collabsphere/internal/uploads/domain"
	"github.com/google/uuid"
)

type GetUploadQuery struct {
	UploadID uuid.UUID
	Actor    authdomain.Principal
}

type Service struct {
	repo        uploadports.Repository
	memberships memberports.MembershipRepository
}

func New(repo uploadports.Repository, memberships memberports.MembershipRepository) *Service {
	return &Service{repo: repo, memberships: memberships}
}

func (s *Service) GetUpload(ctx context.Context, q GetUploadQuery) (*uploaddomain.Upload, error) {
	if q.UploadID == uuid.Nil {
		return nil, fault.Validation("Upload ID is required")
	}
	if !q.Actor.IsAccount() {
		return nil, fault.Unauthorized("Account authentication required")
	}
	if s.repo == nil {
		return nil, fault.Unavailable("Upload tracking is unavailable")
	}

	upload, err := s.repo.GetByID(ctx, q.UploadID)
	if err != nil {
		return nil, fault.Internal("Load upload failed", fault.WithCause(err))
	}
	if upload == nil {
		return nil, fault.NotFound("Upload not found")
	}
	allowed, err := s.accountCanAccessUpload(ctx, upload, q.Actor.AccountID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fault.Forbidden("Upload access denied")
	}
	return upload, nil
}

func (s *Service) accountCanAccessUpload(ctx context.Context, upload *uploaddomain.Upload, accountID uuid.UUID) (bool, error) {
	if upload == nil || accountID == uuid.Nil {
		return false, nil
	}
	if upload.CreatedByAccountID == accountID {
		return true, nil
	}
	if upload.OrganizationID == nil || *upload.OrganizationID == uuid.Nil || s.memberships == nil {
		return false, nil
	}
	orgID, err := orgdomain.OrganizationIDFromUUID(*upload.OrganizationID)
	if err != nil {
		return false, nil
	}
	accountDomainID, err := accdomain.AccountIDFromUUID(accountID)
	if err != nil {
		return false, fault.Unauthorized("Account authentication required")
	}
	membership, err := s.memberships.GetMemberByAccount(ctx, orgID, accountDomainID)
	if err != nil {
		return false, fault.Internal("Resolve upload membership failed", fault.WithCause(err))
	}
	return membership != nil && membership.IsActive() && !membership.IsRemoved(), nil
}
