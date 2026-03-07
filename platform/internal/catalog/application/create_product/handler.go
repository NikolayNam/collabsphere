package create_product

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

func (h *Handler) Handle(ctx context.Context, cmd Command) (*catalogdomain.Product, error) {
	if err := catalogaccess.RequireOrganizationAccess(ctx, h.organizations, h.memberships, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if cmd.CategoryID != nil {
		category, err := h.repo.GetProductCategoryByID(ctx, cmd.OrganizationID, *cmd.CategoryID)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, catalogerrors.ProductCategoryNotFound()
		}
	}

	product, err := catalogdomain.NewProduct(catalogdomain.NewProductParams{
		ID:             catalogdomain.NewProductID(),
		OrganizationID: cmd.OrganizationID,
		CategoryID:     cmd.CategoryID,
		Name:           cmd.Name,
		Description:    cmd.Description,
		SKU:            cmd.SKU,
		PriceAmount:    cmd.PriceAmount,
		CurrencyCode:   cmd.CurrencyCode,
		IsActive:       cmd.IsActive,
		Now:            h.clock.Now(),
	})
	if err != nil {
		return nil, catalogerrors.InvalidInput("Invalid product data")
	}

	if err := h.repo.CreateProduct(ctx, product); err != nil {
		if stdErrors.Is(err, catalogerrors.ErrConflict) {
			return nil, catalogerrors.ProductAlreadyExists()
		}
		return nil, err
	}
	return product, nil
}
