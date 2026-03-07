package application

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/organizations/application/create_organization"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/errors"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/get_organization_by_id"
	"github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
	"github.com/NikolayNam/collabsphere/internal/organizations/domain"
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

func New(repo ports.OrganizationRepository, clock ports.Clock) *Service {
	return &Service{
		create:  create_organization.NewHandler(repo, clock),
		getById: get_organization_by_id.NewHandler(repo)}
}

func (s *Service) CreateOrganization(ctx context.Context, cmd CreateOrganizationCmd) (*domain.Organization, error) {
	return s.create.Handle(ctx, cmd)
}

func (s *Service) GetOrganizationById(ctx context.Context, q GetOrganizationByIdQuery) (*domain.Organization, error) {
	return s.getById.Handle(ctx, q)
}
