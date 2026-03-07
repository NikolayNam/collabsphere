package list_members

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/memberships/application/errors"
	"github.com/NikolayNam/collabsphere/internal/memberships/application/ports"

	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type Handler struct {
	repo      ports.MembershipRepository
	orgReader ports.OrganizationReader
}

func NewHandler(repo ports.MembershipRepository, orgReader ports.OrganizationReader) *Handler {
	return &Handler{repo: repo, orgReader: orgReader}
}

func (h *Handler) Handle(ctx context.Context, orgID orgDomain.OrganizationID) ([]memberDomain.MemberView, error) {
	if orgID.IsZero() {
		return nil, errors.InvalidInput("Invalid organization_id")
	}

	exists, err := h.orgReader.Exists(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.OrganizationNotFound()
	}

	return h.repo.ListMembers(ctx, orgID)
}
