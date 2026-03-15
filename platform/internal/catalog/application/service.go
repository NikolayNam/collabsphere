package application

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	catalogaccess "github.com/NikolayNam/collabsphere/internal/catalog/application/access"
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
	memberports "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	uploadports "github.com/NikolayNam/collabsphere/internal/uploads/application/ports"
	uploaddomain "github.com/NikolayNam/collabsphere/internal/uploads/domain"
	sharedtx "github.com/NikolayNam/collabsphere/shared/tx"
	"github.com/google/uuid"
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

type CompleteProductImportUploadCmd struct {
	OrganizationID orgdomain.OrganizationID
	ActorAccountID accdomain.AccountID
	UploadID       uuid.UUID
	Mode           *string
}

type ProductVideoView struct {
	ID          uuid.UUID
	ObjectID    uuid.UUID
	FileName    string
	ContentType *string
	SizeBytes   int64
	CreatedAt   time.Time
	UploadedBy  *uuid.UUID
	SortOrder   int64
}

type UploadProductVideoCmd struct {
	OrganizationID orgdomain.OrganizationID
	ActorAccountID accdomain.AccountID
	ProductID      catalogdomain.ProductID
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
	repo               ports.CatalogRepository
	organizations      ports.OrganizationReader
	memberships        ports.MembershipReader
	roleResolver      memberports.RoleResolver
	tx                 sharedtx.Manager
	clock              ports.Clock
	storage            ports.ObjectStorage
	storageBucket      string
	uploads            uploadports.Repository
}

func New(repo ports.CatalogRepository, organizations ports.OrganizationReader, memberships ports.MembershipReader, roleResolver memberports.RoleResolver, tx sharedtx.Manager, clock ports.Clock, storage ports.ObjectStorage, storageBucket string, uploads uploadports.Repository) *Service {
	return &Service{
		createCategory:     create_product_category.NewHandler(repo, organizations, memberships, roleResolver, clock),
		updateCategory:     update_product_category.NewHandler(repo, organizations, memberships, roleResolver, clock),
		deleteCategory:     delete_product_category.NewHandler(repo, organizations, memberships, roleResolver, clock),
		listCategories:     list_product_categories.NewHandler(repo, organizations, memberships, roleResolver),
		createProduct:      create_product.NewHandler(repo, organizations, memberships, roleResolver, clock),
		updateProduct:      update_product.NewHandler(repo, organizations, memberships, roleResolver, clock),
		deleteProduct:      delete_product.NewHandler(repo, organizations, memberships, roleResolver, clock),
		listProducts:       list_products.NewHandler(repo, organizations, memberships, roleResolver),
		getProductByID:     get_product_by_id.NewHandler(repo, organizations, memberships, roleResolver),
		createImportUpload: create_product_import_upload.NewHandler(repo, organizations, memberships, roleResolver, clock, storage, storageBucket),
		runImport:          run_product_import.NewHandler(repo, organizations, memberships, roleResolver, clock, storage),
		getImport:          get_product_import.NewHandler(repo, organizations, memberships, roleResolver),
		repo:               repo,
		organizations:      organizations,
		memberships:        memberships,
		roleResolver:       roleResolver,
		tx:                 tx,
		clock:              clock,
		storage:            storage,
		storageBucket:      strings.TrimSpace(storageBucket),
		uploads:            uploads,
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
	if s.createImportUpload == nil || s.uploads == nil || s.tx == nil {
		return nil, catalogerrors.ProductImportUnavailable()
	}
	createdAt := s.clock.Now()
	var result *CreateProductImportUploadResult
	err := s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		uploadResult, err := s.createImportUpload.Handle(ctx, cmd)
		if err != nil {
			return err
		}
		organizationID := cmd.OrganizationID.UUID()
		upload := &uploaddomain.Upload{
			ID:                 uuid.New(),
			OrganizationID:     &organizationID,
			ObjectID:           uploadResult.ObjectID,
			CreatedByAccountID: cmd.ActorAccountID.UUID(),
			Purpose:            uploaddomain.PurposeProductImport,
			Status:             uploaddomain.StatusPending,
			Bucket:             uploadResult.Bucket,
			ObjectKey:          uploadResult.ObjectKey,
			FileName:           uploadResult.FileName,
			ContentType:        cloneOptionalString(cmd.ContentType),
			DeclaredSizeBytes:  uploadResult.SizeBytes,
			ChecksumSHA256:     cloneOptionalString(cmd.ChecksumSHA256),
			Metadata:           map[string]any{},
			ExpiresAt:          &uploadResult.ExpiresAt,
			CreatedAt:          createdAt,
		}
		if err := s.uploads.Create(ctx, upload); err != nil {
			return err
		}
		uploadResult.UploadID = upload.ID
		result = uploadResult
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
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
	upload, err := s.CreateProductImportUpload(ctx, create_product_import_upload.Command{
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

	return s.CompleteProductImportUpload(ctx, CompleteProductImportUploadCmd{
		OrganizationID: cmd.OrganizationID,
		ActorAccountID: cmd.ActorAccountID,
		UploadID:       upload.UploadID,
		Mode:           nil,
	})
}
func (s *Service) UploadProductVideo(ctx context.Context, cmd UploadProductVideoCmd) (*ProductVideoView, error) {
	if cmd.Body == nil {
		return nil, catalogerrors.InvalidInput("Product video file is required")
	}
	if cmd.SizeBytes < 0 {
		return nil, catalogerrors.InvalidInput("Product video size must be non-negative")
	}
	if s.storage == nil || s.storageBucket == "" || s.repo == nil || s.organizations == nil || s.memberships == nil || s.clock == nil {
		return nil, catalogerrors.Internal("product video upload is unavailable", io.ErrClosedPipe)
	}
	if err := catalogaccess.RequireOrganizationEmployeeAccess(ctx, s.organizations, s.memberships, s.roleResolver, cmd.OrganizationID, cmd.ActorAccountID); err != nil {
		return nil, err
	}

	product, err := s.repo.GetProductByID(ctx, cmd.OrganizationID, cmd.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, catalogerrors.ProductNotFound()
	}

	fileName := strings.TrimSpace(cmd.FileName)
	if fileName == "" {
		return nil, catalogerrors.InvalidInput("Product video fileName is required")
	}
	contentType, err := normalizeProductVideoContentType(fileName, cmd.ContentType)
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()
	objectID := uuid.New()
	objectKey := buildProductVideoObjectKey(cmd.OrganizationID.UUID(), cmd.ProductID.UUID(), objectID, fileName, now)
	object := &ports.StorageObject{
		ID:             objectID,
		OrganizationID: cmd.OrganizationID,
		Bucket:         s.storageBucket,
		ObjectKey:      objectKey,
		FileName:       fileName,
		ContentType:    &contentType,
		SizeBytes:      cmd.SizeBytes,
		CreatedAt:      now,
	}
	if err := s.repo.CreateStorageObject(ctx, object); err != nil {
		return nil, err
	}
	if err := s.storage.PutObject(ctx, object.Bucket, object.ObjectKey, cmd.Body, cmd.SizeBytes, contentType); err != nil {
		return nil, catalogerrors.Internal("upload product video", err)
	}
	actorUUID := cmd.ActorAccountID.UUID()
	record, err := s.repo.CreateProductVideo(ctx, cmd.OrganizationID.UUID(), cmd.ProductID.UUID(), objectID, &actorUUID, now)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, catalogerrors.Internal("load product video after create", io.ErrUnexpectedEOF)
	}
	return &ProductVideoView{
		ID:          record.ID,
		ObjectID:    record.ObjectID,
		FileName:    record.FileName,
		ContentType: record.ContentType,
		SizeBytes:   record.SizeBytes,
		CreatedAt:   record.CreatedAt,
		UploadedBy:  record.UploadedBy,
		SortOrder:   record.SortOrder,
	}, nil
}

func (s *Service) ListProductVideos(ctx context.Context, organizationID orgdomain.OrganizationID, productID catalogdomain.ProductID, actorAccountID accdomain.AccountID) ([]ProductVideoView, error) {
	if err := catalogaccess.RequireOrganizationAccess(ctx, s.organizations, s.memberships, s.roleResolver, organizationID, actorAccountID, false); err != nil {
		return nil, err
	}
	items, err := s.repo.ListProductVideos(ctx, organizationID.UUID(), productID.UUID())
	if err != nil {
		return nil, err
	}
	out := make([]ProductVideoView, 0, len(items))
	for _, item := range items {
		out = append(out, ProductVideoView{
			ID:          item.ID,
			ObjectID:    item.ObjectID,
			FileName:    item.FileName,
			ContentType: item.ContentType,
			SizeBytes:   item.SizeBytes,
			CreatedAt:   item.CreatedAt,
			UploadedBy:  item.UploadedBy,
			SortOrder:   item.SortOrder,
		})
	}
	return out, nil
}

func (s *Service) ListProductVideoObjectIDs(ctx context.Context, organizationID orgdomain.OrganizationID, productID catalogdomain.ProductID) ([]uuid.UUID, error) {
	if s.repo == nil {
		return nil, nil
	}
	return s.repo.ListProductVideoObjectIDs(ctx, organizationID.UUID(), productID.UUID())
}

func (s *Service) ListProductVideoObjectIDsByProduct(ctx context.Context, organizationID orgdomain.OrganizationID, productIDs []catalogdomain.ProductID) (map[uuid.UUID][]uuid.UUID, error) {
	if s.repo == nil {
		return map[uuid.UUID][]uuid.UUID{}, nil
	}
	ids := make([]uuid.UUID, 0, len(productIDs))
	for _, productID := range productIDs {
		if productID.IsZero() {
			continue
		}
		ids = append(ids, productID.UUID())
	}
	return s.repo.ListProductVideoObjectIDsByProduct(ctx, organizationID.UUID(), ids)
}

func buildProductVideoObjectKey(organizationID, productID, objectID uuid.UUID, fileName string, now time.Time) string {
	safeName := sanitizeCatalogFileName(fileName, "product-video.mp4")
	return strings.Join([]string{
		"catalog",
		"products",
		organizationID.String(),
		productID.String(),
		"videos",
		now.UTC().Format("2006"),
		now.UTC().Format("01"),
		now.UTC().Format("02"),
		objectID.String(),
		safeName,
	}, "/")
}

func sanitizeCatalogFileName(fileName, fallback string) string {
	base := filepath.Base(strings.TrimSpace(fileName))
	if base == "." || base == string(filepath.Separator) || base == "" {
		return fallback
	}

	var b strings.Builder
	b.Grow(len(base))
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}

	out := strings.Trim(strings.TrimSpace(b.String()), "-")
	if out == "" {
		return fallback
	}
	return out
}

func normalizeProductVideoContentType(fileName, contentType string) (string, error) {
	trimmedType := strings.TrimSpace(contentType)
	if strings.HasPrefix(strings.ToLower(trimmedType), "video/") {
		return trimmedType, nil
	}

	switch strings.ToLower(filepath.Ext(strings.TrimSpace(fileName))) {
	case ".mp4":
		return "video/mp4", nil
	case ".webm":
		return "video/webm", nil
	case ".mov":
		return "video/quicktime", nil
	case ".m4v":
		return "video/x-m4v", nil
	case ".ogv":
		return "video/ogg", nil
	default:
		return "", catalogerrors.InvalidInput("Only product video files are supported")
	}
}
