package create_product_category

import (
	"context"
	stdErrors "errors"

	catalogaccess "github.com/NikolayNam/collabsphere/internal/catalog/application/access"
	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
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

func (h *Handler) Handle(ctx context.Context, cmd Command) (*catalogdomain.ProductCategory, error) {
	if err := catalogaccess.RequireOrganizationAccess(ctx, h.organizations, h.memberships, h.roleResolver, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if cmd.ParentID != nil {
		parent, err := h.repo.GetProductCategoryByID(ctx, cmd.OrganizationID, *cmd.ParentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, catalogerrors.ProductCategoryNotFound()
		}
	}
	status := ""
	if cmd.Status != nil {
		status = *cmd.Status
	}

	category, err := catalogdomain.NewProductCategory(catalogdomain.NewProductCategoryParams{
		ID:             catalogdomain.NewProductCategoryID(),
		OrganizationID: cmd.OrganizationID,
		ParentID:       cmd.ParentID,
		Status:         status,
		Code:           cmd.Code,
		Name:           cmd.Name,
		SortOrder:      cmd.SortOrder,
		Now:            h.clock.Now(),
	})
	if err != nil {
		return nil, catalogerrors.InvalidInput("Invalid product category data")
	}

	if err := h.repo.CreateProductCategory(ctx, category); err != nil {
		if stdErrors.Is(err, catalogerrors.ErrConflict) {
			return nil, catalogerrors.ProductCategoryAlreadyExists()
		}
		return nil, err
	}
	return category, nil
}
