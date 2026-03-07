package update_product

import (
	"context"
	stdErrors "errors"
	"strings"

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

	current, err := h.repo.GetProductByID(ctx, cmd.OrganizationID, cmd.ProductID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, catalogerrors.ProductNotFound()
	}

	categoryID := current.CategoryID()
	if cmd.CategoryID != nil {
		if !cmd.CategoryID.IsZero() {
			category, err := h.repo.GetProductCategoryByID(ctx, cmd.OrganizationID, *cmd.CategoryID)
			if err != nil {
				return nil, err
			}
			if category == nil {
				return nil, catalogerrors.ProductCategoryNotFound()
			}
		}
		categoryID = cmd.CategoryID
	}

	name := current.Name()
	if cmd.Name != nil {
		name = strings.TrimSpace(*cmd.Name)
	}
	description := current.Description()
	if cmd.Description != nil {
		description = cmd.Description
	}
	sku := current.SKU()
	if cmd.SKU != nil {
		sku = cmd.SKU
	}
	priceAmount := current.PriceAmount()
	if cmd.PriceAmount != nil {
		priceAmount = cmd.PriceAmount
	}
	currencyCode := current.CurrencyCode()
	if cmd.CurrencyCode != nil {
		currencyCode = cmd.CurrencyCode
	}
	isActive := current.IsActive()
	if cmd.IsActive != nil {
		isActive = *cmd.IsActive
	}

	updated, err := catalogdomain.RehydrateProduct(catalogdomain.RehydrateProductParams{
		ID:             current.ID(),
		OrganizationID: current.OrganizationID(),
		CategoryID:     categoryID,
		Name:           name,
		Description:    description,
		SKU:            sku,
		PriceAmount:    priceAmount,
		CurrencyCode:   currencyCode,
		IsActive:       isActive,
		CreatedAt:      current.CreatedAt(),
		UpdatedAt:      h.clock.Now(),
	})
	if err != nil {
		return nil, catalogerrors.InvalidInput("Invalid product data")
	}

	if err := h.repo.UpdateProduct(ctx, updated); err != nil {
		if stdErrors.Is(err, catalogerrors.ErrConflict) {
			return nil, catalogerrors.ProductAlreadyExists()
		}
		return nil, err
	}
	return updated, nil
}
