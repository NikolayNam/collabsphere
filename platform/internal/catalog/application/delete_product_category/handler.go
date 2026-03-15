package delete_product_category

import (
	"context"

	catalogaccess "github.com/NikolayNam/collabsphere/internal/catalog/application/access"
	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	memberports "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
)

type Handler struct {
	repo          ports.CatalogRepository
	organizations ports.OrganizationReader
	memberships   ports.MembershipReader
	roleResolver  memberports.RoleResolver
	clock         ports.Clock
}

func NewHandler(repo ports.CatalogRepository, organizations ports.OrganizationReader, memberships ports.MembershipReader, roleResolver memberports.RoleResolver, clock ports.Clock) *Handler {
	return &Handler{repo: repo, organizations: organizations, memberships: memberships, roleResolver: roleResolver, clock: clock}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	if err := catalogaccess.RequireOrganizationAccess(ctx, h.organizations, h.memberships, h.roleResolver, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return err
	}
	current, err := h.repo.GetProductCategoryByID(ctx, cmd.OrganizationID, cmd.CategoryID)
	if err != nil {
		return err
	}
	if current == nil {
		return catalogerrors.ProductCategoryNotFound()
	}
	return h.repo.DeleteProductCategory(ctx, cmd.OrganizationID, cmd.CategoryID, h.clock.Now())
}
