package application

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/memberships/application/add_member"
	"github.com/NikolayNam/collabsphere/internal/memberships/application/errors"

	"github.com/NikolayNam/collabsphere/internal/memberships/application/list_members"

	memberPorts "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
	orgPorts "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"

	memberDomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	orgDomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

var (
	ErrValidation = errors.ErrValidation
)

type orgReaderAdapter struct {
	repo orgPorts.OrganizationRepository
}

func (a orgReaderAdapter) Exists(ctx context.Context, id orgDomain.OrganizationID) (bool, error) {
	org, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	return org != nil, nil
}

type Service struct {
	addMember   *add_member.Handler
	listMembers *list_members.Handler
}

func New(memberRepo memberPorts.MembershipRepository, orgRepo orgPorts.OrganizationRepository, clock memberPorts.Clock) *Service {
	orgReader := orgReaderAdapter{repo: orgRepo}

	return &Service{
		addMember:   add_member.NewHandler(memberRepo, orgReader, clock),
		listMembers: list_members.NewHandler(memberRepo, orgReader),
	}
}

func (s *Service) AddMember(ctx context.Context, orgID orgDomain.OrganizationID, accountID string, kind string) error {
	return s.addMember.Handle(ctx, orgID, accountID, kind)
}

func (s *Service) ListMembers(ctx context.Context, orgID orgDomain.OrganizationID) ([]memberDomain.MemberView, error) {
	return s.listMembers.Handle(ctx, orgID)
}
