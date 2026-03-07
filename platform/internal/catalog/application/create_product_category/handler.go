package create_product_category

import (
	"context"
	stdErrors "errors"

	catalogaccess "github.com/NikolayNam/collabsphere/internal/catalog/application/access"
	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
)

type Handler struct {
	repo          ports.CatalogRepository
	organizations ports.OrganizationReader
	memberships   ports.MembershipReader
	clock         ports.Clock
}

func NewHandler(repo ports.CatalogRepository, organizations ports.OrganizationReader, memberships ports.MembershipReader, clock ports.Clock) *Handler {
	return &Handler{repo: repo, organizations: organizations, memberships: memberships, clock: clock}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*catalogdomain.ProductCategory, error) {
	if err := catalogaccess.RequireOrganizationAccess(ctx, h.organizations, h.memberships, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
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

	category, err := catalogdomain.NewProductCategory(catalogdomain.NewProductCategoryParams{
		ID:             catalogdomain.NewProductCategoryID(),
		OrganizationID: cmd.OrganizationID,
		ParentID:       cmd.ParentID,
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
