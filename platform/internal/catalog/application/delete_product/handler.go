package delete_product

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
	current, err := h.repo.GetProductByID(ctx, cmd.OrganizationID, cmd.ProductID)
	if err != nil {
		return err
	}
	if current == nil {
		return catalogerrors.ProductNotFound()
	}
	return h.repo.DeleteProduct(ctx, cmd.OrganizationID, cmd.ProductID, h.clock.Now())
}
