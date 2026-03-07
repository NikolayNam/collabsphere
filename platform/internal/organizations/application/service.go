package application

import (
	"context"

	memberPorts "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/create_organization"
	create_with_owner "github.com/NikolayNam/collabsphere/internal/organizations/application/create_organization_with_owner"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/get_organization_by_id"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
	sharedtx "github.com/NikolayNam/collabsphere/shared/tx"
)

var (
	ErrValidation = errors.ErrValidation
)

type CreateOrganizationCmd = create_organization.Command
type GetOrganizationByIdQuery = get_organization_by_id.Query

type Service struct {
	create  *create_organization.Handler
	getById *get_organization_by_id.Handler
}

func New(repo ports.OrganizationRepository, membershipRepo memberPorts.MembershipRepository, txm sharedtx.Manager, clock ports.Clock) *Service {
	creator := create_with_owner.New(txm, repo, membershipRepo)

	return &Service{
		create:  create_organization.NewHandler(creator, clock),
		getById: get_organization_by_id.NewHandler(repo),
	}
}

func (s *Service) CreateOrganization(ctx context.Context, cmd CreateOrganizationCmd) (*domain.Organization, error) {
	return s.create.Handle(ctx, cmd)
}

func (s *Service) GetOrganizationById(ctx context.Context, q GetOrganizationByIdQuery) (*domain.Organization, error) {
	return s.getById.Handle(ctx, q)
}
