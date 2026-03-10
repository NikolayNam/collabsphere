package ports

import (
	"context"
	"time"

	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
)

type ProductVideoRecord struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	ProductID      uuid.UUID
	ObjectID       uuid.UUID
	FileName       string
	ContentType    *string
	SizeBytes      int64
	CreatedAt      time.Time
	UploadedBy     *uuid.UUID
	SortOrder      int64
}

type CatalogRepository interface {
	CreateProductCategory(ctx context.Context, category *catalogdomain.ProductCategory) error
	GetProductCategoryByID(ctx context.Context, organizationID orgdomain.OrganizationID, categoryID catalogdomain.ProductCategoryID) (*catalogdomain.ProductCategory, error)
	FindProductCategoryByCode(ctx context.Context, organizationID orgdomain.OrganizationID, code string) (*catalogdomain.ProductCategory, error)
	ListProductCategories(ctx context.Context, organizationID orgdomain.OrganizationID) ([]catalogdomain.ProductCategory, error)
	UpdateProductCategory(ctx context.Context, category *catalogdomain.ProductCategory) error
	DeleteProductCategory(ctx context.Context, organizationID orgdomain.OrganizationID, categoryID catalogdomain.ProductCategoryID, deletedAt time.Time) error

	CreateProduct(ctx context.Context, product *catalogdomain.Product) error
	GetProductByID(ctx context.Context, organizationID orgdomain.OrganizationID, productID catalogdomain.ProductID) (*catalogdomain.Product, error)
	GetProductBySKU(ctx context.Context, organizationID orgdomain.OrganizationID, sku string) (*catalogdomain.Product, error)
	ListProducts(ctx context.Context, organizationID orgdomain.OrganizationID) ([]catalogdomain.Product, error)
	UpdateProduct(ctx context.Context, product *catalogdomain.Product) error
	DeleteProduct(ctx context.Context, organizationID orgdomain.OrganizationID, productID catalogdomain.ProductID, deletedAt time.Time) error
	CreateProductVideo(ctx context.Context, organizationID, productID, objectID uuid.UUID, uploadedBy *uuid.UUID, createdAt time.Time) (*ProductVideoRecord, error)
	ListProductVideos(ctx context.Context, organizationID, productID uuid.UUID) ([]ProductVideoRecord, error)
	ListProductVideoObjectIDs(ctx context.Context, organizationID, productID uuid.UUID) ([]uuid.UUID, error)
	ListProductVideoObjectIDsByProduct(ctx context.Context, organizationID uuid.UUID, productIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error)

	CreateStorageObject(ctx context.Context, object *StorageObject) error
	GetStorageObjectByID(ctx context.Context, organizationID orgdomain.OrganizationID, objectID uuid.UUID) (*StorageObject, error)

	CreateProductImportBatch(ctx context.Context, batch *ProductImportBatch) error
	UpdateProductImportBatch(ctx context.Context, batch *ProductImportBatch) error
	GetProductImportBatchByID(ctx context.Context, organizationID orgdomain.OrganizationID, batchID uuid.UUID) (*ProductImportBatch, error)
	GetProductImportBatchBySourceObjectID(ctx context.Context, organizationID orgdomain.OrganizationID, sourceObjectID uuid.UUID) (*ProductImportBatch, error)
	AddProductImportErrors(ctx context.Context, batchID uuid.UUID, items []ProductImportErrorRecord) error
	ListProductImportErrors(ctx context.Context, batchID uuid.UUID) ([]ProductImportErrorRecord, error)
}
