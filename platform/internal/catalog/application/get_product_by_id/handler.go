package get_product_by_id

import (
	"context"

	catalogaccess "github.com/NikolayNam/collabsphere/internal/catalog/application/access"
	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
)

type Handler struct {
	repo          ports.CatalogRepository
	organizations ports.OrganizationReader
	memberships   ports.MembershipReader
}

func NewHandler(repo ports.CatalogRepository, organizations ports.OrganizationReader, memberships ports.MembershipReader) *Handler {
	return &Handler{repo: repo, organizations: organizations, memberships: memberships}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*catalogdomain.Product, error) {
	if err := catalogaccess.RequireOrganizationAccess(ctx, h.organizations, h.memberships, q.OrganizationID, q.ActorAccountID, false); err != nil {
		return nil, err
	}
	product, err := h.repo.GetProductByID(ctx, q.OrganizationID, q.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, catalogerrors.ProductNotFound()
	}
	return product, nil
}
