package get_product_import

import (
	"context"

	catalogaccess "github.com/NikolayNam/collabsphere/internal/catalog/application/access"
	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	productimport "github.com/NikolayNam/collabsphere/internal/catalog/application/product_import"
)

type Handler struct {
	repo          ports.CatalogRepository
	organizations ports.OrganizationReader
	memberships   ports.MembershipReader
}

func NewHandler(repo ports.CatalogRepository, organizations ports.OrganizationReader, memberships ports.MembershipReader) *Handler {
	return &Handler{repo: repo, organizations: organizations, memberships: memberships}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*productimport.View, error) {
	if err := catalogaccess.RequireOrganizationAccess(ctx, h.organizations, h.memberships, q.OrganizationID, q.ActorAccountID, false); err != nil {
		return nil, err
	}

	batch, err := h.repo.GetProductImportBatchByID(ctx, q.OrganizationID, q.BatchID)
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, catalogerrors.ProductImportNotFound()
	}

	items, err := h.repo.ListProductImportErrors(ctx, q.BatchID)
	if err != nil {
		return nil, err
	}

	return &productimport.View{Batch: batch, Errors: items}, nil
}
