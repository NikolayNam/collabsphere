package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxCatalogListRows = 500

type catalogProductCategoryRow struct {
	ID             uuid.UUID  `gorm:"column:id"`
	OrganizationID uuid.UUID  `gorm:"column:organization_id"`
	ParentID       *uuid.UUID `gorm:"column:parent_id"`
	TemplateID     *uuid.UUID `gorm:"column:template_id"`
	Code           string     `gorm:"column:code"`
	Name           string     `gorm:"column:name"`
	SortOrder      int64      `gorm:"column:sort_order"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

type catalogProductRow struct {
	ID             uuid.UUID  `gorm:"column:id"`
	OrganizationID uuid.UUID  `gorm:"column:organization_id"`
	CategoryID     *uuid.UUID `gorm:"column:product_type_id"`
	Name           string     `gorm:"column:name"`
	Description    *string    `gorm:"column:description"`
	SKU            *string    `gorm:"column:sku"`
	PriceAmount    *string    `gorm:"column:price_amount"`
	CurrencyCode   *string    `gorm:"column:currency_code"`
	IsActive       bool       `gorm:"column:is_active"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

type storageObjectRow struct {
	ID             uuid.UUID  `gorm:"column:id"`
	OrganizationID uuid.UUID  `gorm:"column:organization_id"`
	Bucket         string     `gorm:"column:bucket"`
	ObjectKey      string     `gorm:"column:object_key"`
	FileName       string     `gorm:"column:file_name"`
	ContentType    *string    `gorm:"column:content_type"`
	SizeBytes      int64      `gorm:"column:size_bytes"`
	ChecksumSHA256 *string    `gorm:"column:checksum_sha256"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

type productImportBatchRow struct {
	ID                 uuid.UUID  `gorm:"column:id"`
	OrganizationID     uuid.UUID  `gorm:"column:organization_id"`
	SourceObjectID     uuid.UUID  `gorm:"column:source_object_id"`
	CreatedByAccountID uuid.UUID  `gorm:"column:created_by_account_id"`
	Status             string     `gorm:"column:status"`
	TotalRows          *int       `gorm:"column:total_rows"`
	ProcessedRows      int        `gorm:"column:processed_rows"`
	SuccessRows        int        `gorm:"column:success_rows"`
	ErrorRows          int        `gorm:"column:error_rows"`
	StartedBy          *string    `gorm:"column:started_by"`
	StartedAt          time.Time  `gorm:"column:started_at"`
	FinishedAt         *time.Time `gorm:"column:finished_at"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          *time.Time `gorm:"column:updated_at"`
	Mode               *string    `gorm:"column:mode"`
	ResultSummary      *string    `gorm:"column:result_summary"`
}

type productImportErrorRow struct {
	ID        uuid.UUID `gorm:"column:id"`
	BatchID   uuid.UUID `gorm:"column:batch_id"`
	RowNo     *int      `gorm:"column:row_no"`
	Code      *string   `gorm:"column:code"`
	Message   string    `gorm:"column:message"`
	Details   *string   `gorm:"column:details"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (r *CatalogRepo) GetProductCategoryByID(ctx context.Context, organizationID orgdomain.OrganizationID, categoryID catalogdomain.ProductCategoryID) (*catalogdomain.ProductCategory, error) {
	var m catalogProductCategoryRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.product_categories").
		Select("id", "organization_id", "parent_id", "template_id", "code", "name", "sort_order", "created_at", "updated_at").
		Where("organization_id = ? AND id = ? AND deleted_at IS NULL", organizationID.UUID(), categoryID.UUID()).
		Take(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return rehydrateCategoryRow(m)
}

func (r *CatalogRepo) FindProductCategoryByCode(ctx context.Context, organizationID orgdomain.OrganizationID, code string) (*catalogdomain.ProductCategory, error) {
	var m catalogProductCategoryRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.product_categories").
		Select("id", "organization_id", "parent_id", "template_id", "code", "name", "sort_order", "created_at", "updated_at").
		Where("organization_id = ? AND code = ? AND deleted_at IS NULL", organizationID.UUID(), strings.TrimSpace(code)).
		Take(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rehydrateCategoryRow(m)
}

func (r *CatalogRepo) ListProductCategories(ctx context.Context, organizationID orgdomain.OrganizationID) ([]catalogdomain.ProductCategory, error) {
	var rows []catalogProductCategoryRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.product_categories").
		Select("id", "organization_id", "parent_id", "template_id", "code", "name", "sort_order", "created_at", "updated_at").
		Where("organization_id = ? AND deleted_at IS NULL", organizationID.UUID()).
		Order("sort_order ASC, name ASC, id ASC").
		Limit(maxCatalogListRows).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]catalogdomain.ProductCategory, 0, len(rows))
	for _, row := range rows {
		category, err := rehydrateCategoryRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, *category)
	}
	return out, nil
}

func (r *CatalogRepo) GetProductByID(ctx context.Context, organizationID orgdomain.OrganizationID, productID catalogdomain.ProductID) (*catalogdomain.Product, error) {
	var m catalogProductRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.products").
		Select("id", "organization_id", "product_type_id", "name", "description", "sku", "price_amount::text AS price_amount", "currency_code", "is_active", "created_at", "updated_at").
		Where("organization_id = ? AND id = ? AND deleted_at IS NULL", organizationID.UUID(), productID.UUID()).
		Take(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return rehydrateProductRow(m)
}

func (r *CatalogRepo) GetProductBySKU(ctx context.Context, organizationID orgdomain.OrganizationID, sku string) (*catalogdomain.Product, error) {
	trimmed := strings.TrimSpace(sku)
	if trimmed == "" {
		return nil, nil
	}

	var rows []catalogProductRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.products").
		Select("id", "organization_id", "product_type_id", "name", "description", "sku", "price_amount::text AS price_amount", "currency_code", "is_active", "created_at", "updated_at").
		Where("organization_id = ? AND sku = ? AND deleted_at IS NULL", organizationID.UUID(), trimmed).
		Order("created_at DESC, id DESC").
		Limit(2).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, nil
	}
	if len(rows) > 1 {
		return nil, fmt.Errorf("multiple products found for sku %q", trimmed)
	}

	return rehydrateProductRow(rows[0])
}

func (r *CatalogRepo) ListProducts(ctx context.Context, organizationID orgdomain.OrganizationID) ([]catalogdomain.Product, error) {
	var rows []catalogProductRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.products").
		Select("id", "organization_id", "product_type_id", "name", "description", "sku", "price_amount::text AS price_amount", "currency_code", "is_active", "created_at", "updated_at").
		Where("organization_id = ? AND deleted_at IS NULL", organizationID.UUID()).
		Order("created_at DESC, id DESC").
		Limit(maxCatalogListRows).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]catalogdomain.Product, 0, len(rows))
	for _, row := range rows {
		product, err := rehydrateProductRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, *product)
	}
	return out, nil
}

func (r *CatalogRepo) GetStorageObjectByID(ctx context.Context, organizationID orgdomain.OrganizationID, objectID uuid.UUID) (*ports.StorageObject, error) {
	var row storageObjectRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("storage.objects").
		Select("id", "organization_id", "bucket", "object_key", "file_name", "content_type", "size_bytes", "checksum_sha256", "created_at", "deleted_at").
		Where("organization_id = ? AND id = ? AND deleted_at IS NULL", organizationID.UUID(), objectID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toStorageObject(row)
}

func (r *CatalogRepo) GetProductImportBatchByID(ctx context.Context, organizationID orgdomain.OrganizationID, batchID uuid.UUID) (*ports.ProductImportBatch, error) {
	var row productImportBatchRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.product_import_batches").
		Select("id", "organization_id", "source_object_id", "created_by_account_id", "status", "total_rows", "processed_rows", "success_rows", "error_rows", "started_by", "started_at", "finished_at", "created_at", "updated_at", "mode", "result_summary::text AS result_summary").
		Where("organization_id = ? AND id = ?", organizationID.UUID(), batchID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toProductImportBatch(row)
}

func (r *CatalogRepo) GetProductImportBatchBySourceObjectID(ctx context.Context, organizationID orgdomain.OrganizationID, sourceObjectID uuid.UUID) (*ports.ProductImportBatch, error) {
	var row productImportBatchRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.product_import_batches").
		Select("id", "organization_id", "source_object_id", "created_by_account_id", "status", "total_rows", "processed_rows", "success_rows", "error_rows", "started_by", "started_at", "finished_at", "created_at", "updated_at", "mode", "result_summary::text AS result_summary").
		Where("organization_id = ? AND source_object_id = ?", organizationID.UUID(), sourceObjectID).
		Order("created_at DESC, id DESC").
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toProductImportBatch(row)
}

func (r *CatalogRepo) ListProductImportErrors(ctx context.Context, batchID uuid.UUID) ([]ports.ProductImportErrorRecord, error) {
	var rows []productImportErrorRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.product_import_errors").
		Select("id", "batch_id", "row_no", "code", "message", "details::text AS details", "created_at").
		Where("batch_id = ?", batchID).
		Order("row_no ASC NULLS FIRST, created_at ASC, id ASC").
		Limit(maxCatalogListRows).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]ports.ProductImportErrorRecord, 0, len(rows))
	for _, row := range rows {
		item, err := toProductImportError(row)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, nil
}

func rehydrateCategoryRow(row catalogProductCategoryRow) (*catalogdomain.ProductCategory, error) {
	organizationID, err := orgdomain.OrganizationIDFromUUID(row.OrganizationID)
	if err != nil {
		return nil, err
	}
	categoryID, err := catalogdomain.ProductCategoryIDFromUUID(row.ID)
	if err != nil {
		return nil, err
	}
	var parentID *catalogdomain.ProductCategoryID
	if row.ParentID != nil {
		parent, err := catalogdomain.ProductCategoryIDFromUUID(*row.ParentID)
		if err != nil {
			return nil, err
		}
		parentID = &parent
	}
	return catalogdomain.RehydrateProductCategory(catalogdomain.RehydrateProductCategoryParams{
		ID:             categoryID,
		OrganizationID: organizationID,
		ParentID:       parentID,
		TemplateID:     row.TemplateID,
		Code:           row.Code,
		Name:           row.Name,
		SortOrder:      row.SortOrder,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	})
}

func rehydrateProductRow(row catalogProductRow) (*catalogdomain.Product, error) {
	organizationID, err := orgdomain.OrganizationIDFromUUID(row.OrganizationID)
	if err != nil {
		return nil, err
	}
	productID, err := catalogdomain.ProductIDFromUUID(row.ID)
	if err != nil {
		return nil, err
	}
	var categoryID *catalogdomain.ProductCategoryID
	if row.CategoryID != nil {
		category, err := catalogdomain.ProductCategoryIDFromUUID(*row.CategoryID)
		if err != nil {
			return nil, err
		}
		categoryID = &category
	}
	return catalogdomain.RehydrateProduct(catalogdomain.RehydrateProductParams{
		ID:             productID,
		OrganizationID: organizationID,
		CategoryID:     categoryID,
		Name:           row.Name,
		Description:    row.Description,
		SKU:            row.SKU,
		PriceAmount:    row.PriceAmount,
		CurrencyCode:   row.CurrencyCode,
		IsActive:       row.IsActive,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	})
}

func toStorageObject(row storageObjectRow) (*ports.StorageObject, error) {
	organizationID, err := orgdomain.OrganizationIDFromUUID(row.OrganizationID)
	if err != nil {
		return nil, err
	}
	return &ports.StorageObject{
		ID:             row.ID,
		OrganizationID: organizationID,
		Bucket:         row.Bucket,
		ObjectKey:      row.ObjectKey,
		FileName:       row.FileName,
		ContentType:    row.ContentType,
		SizeBytes:      row.SizeBytes,
		ChecksumSHA256: row.ChecksumSHA256,
		CreatedAt:      row.CreatedAt,
		DeletedAt:      row.DeletedAt,
	}, nil
}

func toProductImportBatch(row productImportBatchRow) (*ports.ProductImportBatch, error) {
	organizationID, err := orgdomain.OrganizationIDFromUUID(row.OrganizationID)
	if err != nil {
		return nil, err
	}
	accountID, err := accdomain.AccountIDFromUUID(row.CreatedByAccountID)
	if err != nil {
		return nil, err
	}
	return &ports.ProductImportBatch{
		ID:                 row.ID,
		OrganizationID:     organizationID,
		SourceObjectID:     row.SourceObjectID,
		CreatedByAccountID: accountID,
		Status:             ports.ProductImportStatus(strings.TrimSpace(row.Status)),
		TotalRows:          row.TotalRows,
		ProcessedRows:      row.ProcessedRows,
		SuccessRows:        row.SuccessRows,
		ErrorRows:          row.ErrorRows,
		StartedBy:          row.StartedBy,
		StartedAt:          row.StartedAt,
		FinishedAt:         row.FinishedAt,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
		Mode:               row.Mode,
		ResultSummary:      decodeJSONMap(row.ResultSummary),
	}, nil
}

func toProductImportError(row productImportErrorRow) (*ports.ProductImportErrorRecord, error) {
	return &ports.ProductImportErrorRecord{
		ID:        row.ID,
		BatchID:   row.BatchID,
		RowNo:     row.RowNo,
		Code:      row.Code,
		Message:   row.Message,
		Details:   decodeJSONMap(row.Details),
		CreatedAt: row.CreatedAt,
	}, nil
}

func decodeJSONMap(value *string) map[string]any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return map[string]any{}
	}

	out := map[string]any{}
	if err := json.Unmarshal([]byte(*value), &out); err != nil {
		return map[string]any{}
	}
	return out
}
