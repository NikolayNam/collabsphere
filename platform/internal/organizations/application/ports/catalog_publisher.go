package ports

import (
	"context"

	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

// CatalogPublisher provides catalog bulk operations for organization-scoped publish flows.
// Implemented by catalog repository.
type CatalogPublisher interface {
	ListProductCategories(ctx context.Context, organizationID orgdomain.OrganizationID) ([]catalogdomain.ProductCategory, error)
	ListProducts(ctx context.Context, organizationID orgdomain.OrganizationID) ([]catalogdomain.Product, error)
	UpdateProductCategory(ctx context.Context, category *catalogdomain.ProductCategory) error
	UpdateProduct(ctx context.Context, product *catalogdomain.Product) error
}
