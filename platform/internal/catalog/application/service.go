package application

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/catalog/application/create_product"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/create_product_category"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/get_product_by_id"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/list_product_categories"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/list_products"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
)

var (
	ErrValidation = errors.ErrValidation
)

type CreateProductCategoryCmd = create_product_category.Command
type ListProductCategoriesQuery = list_product_categories.Query
type CreateProductCmd = create_product.Command
type ListProductsQuery = list_products.Query
type GetProductByIDQuery = get_product_by_id.Query

type Service struct {
	createCategory *create_product_category.Handler
	listCategories *list_product_categories.Handler
	createProduct  *create_product.Handler
	listProducts   *list_products.Handler
	getProductByID *get_product_by_id.Handler
}

func New(repo ports.CatalogRepository, organizations ports.OrganizationReader, memberships ports.MembershipReader, clock ports.Clock) *Service {
	return &Service{
		createCategory: create_product_category.NewHandler(repo, organizations, memberships, clock),
		listCategories: list_product_categories.NewHandler(repo, organizations, memberships),
		createProduct:  create_product.NewHandler(repo, organizations, memberships, clock),
		listProducts:   list_products.NewHandler(repo, organizations, memberships),
		getProductByID: get_product_by_id.NewHandler(repo, organizations, memberships),
	}
}

func (s *Service) CreateProductCategory(ctx context.Context, cmd CreateProductCategoryCmd) (*catalogdomain.ProductCategory, error) {
	return s.createCategory.Handle(ctx, cmd)
}

func (s *Service) ListProductCategories(ctx context.Context, q ListProductCategoriesQuery) ([]catalogdomain.ProductCategory, error) {
	return s.listCategories.Handle(ctx, q)
}

func (s *Service) CreateProduct(ctx context.Context, cmd CreateProductCmd) (*catalogdomain.Product, error) {
	return s.createProduct.Handle(ctx, cmd)
}

func (s *Service) ListProducts(ctx context.Context, q ListProductsQuery) ([]catalogdomain.Product, error) {
	return s.listProducts.Handle(ctx, q)
}

func (s *Service) GetProductByID(ctx context.Context, q GetProductByIDQuery) (*catalogdomain.Product, error) {
	return s.getProductByID.Handle(ctx, q)
}
