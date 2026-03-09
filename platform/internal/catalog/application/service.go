package application

import (
	"context"
	"io"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/create_product"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/create_product_category"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/create_product_import_upload"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/delete_product"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/delete_product_category"
	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/get_product_by_id"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/get_product_import"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/list_product_categories"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/list_products"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	productimport "github.com/NikolayNam/collabsphere/internal/catalog/application/product_import"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/run_product_import"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/update_product"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/update_product_category"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
)

var (
	ErrValidation = catalogerrors.ErrValidation
)

type CreateProductCategoryCmd = create_product_category.Command
type UpdateProductCategoryCmd = update_product_category.Command
type DeleteProductCategoryCmd = delete_product_category.Command
type ListProductCategoriesQuery = list_product_categories.Query
type CreateProductCmd = create_product.Command
type UpdateProductCmd = update_product.Command
type DeleteProductCmd = delete_product.Command
type ListProductsQuery = list_products.Query
type GetProductByIDQuery = get_product_by_id.Query
type CreateProductImportUploadCmd = create_product_import_upload.Command
type CreateProductImportUploadResult = create_product_import_upload.Result
type RunProductImportCmd = run_product_import.Command
type GetProductImportQuery = get_product_import.Query
type ProductImportView = productimport.View

type UploadProductImportCmd struct {
	OrganizationID orgdomain.OrganizationID
	ActorAccountID accdomain.AccountID
	FileName       string
	ContentType    string
	SizeBytes      int64
	Body           io.Reader
}

type Service struct {
	createCategory     *create_product_category.Handler
	updateCategory     *update_product_category.Handler
	deleteCategory     *delete_product_category.Handler
	listCategories     *list_product_categories.Handler
	createProduct      *create_product.Handler
	updateProduct      *update_product.Handler
	deleteProduct      *delete_product.Handler
	listProducts       *list_products.Handler
	getProductByID     *get_product_by_id.Handler
	createImportUpload *create_product_import_upload.Handler
	runImport          *run_product_import.Handler
	getImport          *get_product_import.Handler
	storage            ports.ObjectStorage
}

func New(repo ports.CatalogRepository, organizations ports.OrganizationReader, memberships ports.MembershipReader, clock ports.Clock, storage ports.ObjectStorage, storageBucket string) *Service {
	return &Service{
		createCategory:     create_product_category.NewHandler(repo, organizations, memberships, clock),
		updateCategory:     update_product_category.NewHandler(repo, organizations, memberships, clock),
		deleteCategory:     delete_product_category.NewHandler(repo, organizations, memberships, clock),
		listCategories:     list_product_categories.NewHandler(repo, organizations, memberships),
		createProduct:      create_product.NewHandler(repo, organizations, memberships, clock),
		updateProduct:      update_product.NewHandler(repo, organizations, memberships, clock),
		deleteProduct:      delete_product.NewHandler(repo, organizations, memberships, clock),
		listProducts:       list_products.NewHandler(repo, organizations, memberships),
		getProductByID:     get_product_by_id.NewHandler(repo, organizations, memberships),
		createImportUpload: create_product_import_upload.NewHandler(repo, organizations, memberships, clock, storage, storageBucket),
		runImport:          run_product_import.NewHandler(repo, organizations, memberships, clock, storage),
		getImport:          get_product_import.NewHandler(repo, organizations, memberships),
		storage:            storage,
	}
}

func (s *Service) CreateProductCategory(ctx context.Context, cmd CreateProductCategoryCmd) (*catalogdomain.ProductCategory, error) {
	return s.createCategory.Handle(ctx, cmd)
}

func (s *Service) UpdateProductCategory(ctx context.Context, cmd UpdateProductCategoryCmd) (*catalogdomain.ProductCategory, error) {
	return s.updateCategory.Handle(ctx, cmd)
}

func (s *Service) DeleteProductCategory(ctx context.Context, cmd DeleteProductCategoryCmd) error {
	return s.deleteCategory.Handle(ctx, cmd)
}

func (s *Service) ListProductCategories(ctx context.Context, q ListProductCategoriesQuery) ([]catalogdomain.ProductCategory, error) {
	return s.listCategories.Handle(ctx, q)
}

func (s *Service) CreateProduct(ctx context.Context, cmd CreateProductCmd) (*catalogdomain.Product, error) {
	return s.createProduct.Handle(ctx, cmd)
}

func (s *Service) UpdateProduct(ctx context.Context, cmd UpdateProductCmd) (*catalogdomain.Product, error) {
	return s.updateProduct.Handle(ctx, cmd)
}

func (s *Service) DeleteProduct(ctx context.Context, cmd DeleteProductCmd) error {
	return s.deleteProduct.Handle(ctx, cmd)
}

func (s *Service) ListProducts(ctx context.Context, q ListProductsQuery) ([]catalogdomain.Product, error) {
	return s.listProducts.Handle(ctx, q)
}

func (s *Service) GetProductByID(ctx context.Context, q GetProductByIDQuery) (*catalogdomain.Product, error) {
	return s.getProductByID.Handle(ctx, q)
}

func (s *Service) CreateProductImportUpload(ctx context.Context, cmd CreateProductImportUploadCmd) (*CreateProductImportUploadResult, error) {
	return s.createImportUpload.Handle(ctx, cmd)
}

func (s *Service) RunProductImport(ctx context.Context, cmd RunProductImportCmd) (*ProductImportView, error) {
	return s.runImport.Handle(ctx, cmd)
}

func (s *Service) GetProductImport(ctx context.Context, q GetProductImportQuery) (*ProductImportView, error) {
	return s.getImport.Handle(ctx, q)
}

func (s *Service) UploadProductImport(ctx context.Context, cmd UploadProductImportCmd) (*ProductImportView, error) {
	if cmd.Body == nil {
		return nil, catalogerrors.ProductImportFileInvalid("file is required")
	}
	if cmd.SizeBytes < 0 {
		return nil, catalogerrors.ProductImportFileInvalid("file size must be non-negative")
	}
	if s.createImportUpload == nil || s.runImport == nil || s.storage == nil {
		return nil, catalogerrors.ProductImportUnavailable()
	}

	contentType := cmd.ContentType
	sizeBytes := cmd.SizeBytes
	upload, err := s.createImportUpload.Handle(ctx, create_product_import_upload.Command{
		OrganizationID: cmd.OrganizationID,
		ActorAccountID: cmd.ActorAccountID,
		FileName:       cmd.FileName,
		ContentType:    &contentType,
		SizeBytes:      &sizeBytes,
	})
	if err != nil {
		return nil, err
	}

	if err := s.storage.PutObject(ctx, upload.Bucket, upload.ObjectKey, cmd.Body, cmd.SizeBytes, cmd.ContentType); err != nil {
		return nil, catalogerrors.Internal("upload product import file", err)
	}

	return s.runImport.Handle(ctx, run_product_import.Command{
		OrganizationID: cmd.OrganizationID,
		ActorAccountID: cmd.ActorAccountID,
		SourceObjectID: upload.ObjectID,
		Mode:           nil,
	})
}
