package ports

import (
	"context"

	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

type CatalogRepository interface {
	CreateProductCategory(ctx context.Context, category *catalogdomain.ProductCategory) error
	GetProductCategoryByID(ctx context.Context, organizationID orgdomain.OrganizationID, categoryID catalogdomain.ProductCategoryID) (*catalogdomain.ProductCategory, error)
	ListProductCategories(ctx context.Context, organizationID orgdomain.OrganizationID) ([]catalogdomain.ProductCategory, error)
	CreateProduct(ctx context.Context, product *catalogdomain.Product) error
	GetProductByID(ctx context.Context, organizationID orgdomain.OrganizationID, productID catalogdomain.ProductID) (*catalogdomain.Product, error)
	ListProducts(ctx context.Context, organizationID orgdomain.OrganizationID) ([]catalogdomain.Product, error)
}
