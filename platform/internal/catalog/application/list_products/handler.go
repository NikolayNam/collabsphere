package list_products

import (
	"context"

	catalogaccess "github.com/NikolayNam/collabsphere/internal/catalog/application/access"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	memberports "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
)

type Handler struct {
	repo          ports.CatalogRepository
	organizations ports.OrganizationReader
	memberships   ports.MembershipReader
	roleResolver  memberports.RoleResolver
}

func NewHandler(repo ports.CatalogRepository, organizations ports.OrganizationReader, memberships ports.MembershipReader, roleResolver memberports.RoleResolver) *Handler {
	return &Handler{repo: repo, organizations: organizations, memberships: memberships, roleResolver: roleResolver}
}

func (h *Handler) Handle(ctx context.Context, q Query) ([]catalogdomain.Product, error) {
	if err := catalogaccess.RequireOrganizationAccess(ctx, h.organizations, h.memberships, h.roleResolver, q.OrganizationID, q.ActorAccountID, false); err != nil {
		return nil, err
	}
	return h.repo.ListProducts(ctx, q.OrganizationID)
}
