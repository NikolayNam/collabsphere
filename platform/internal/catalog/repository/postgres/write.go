package postgres

import (
	"context"
	"encoding/json"
	"time"

	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *CatalogRepo) CreateProductCategory(ctx context.Context, category *catalogdomain.ProductCategory) error {
	if category == nil {
		return catalogerrors.InvalidInput("Product category is required")
	}

	updatedAt := category.CreatedAt()
	if category.UpdatedAt() != nil {
		updatedAt = *category.UpdatedAt()
	}

	payload := map[string]any{
		"id":              category.ID().UUID(),
		"organization_id": category.OrganizationID().UUID(),
		"parent_id":       nullableProductCategoryID(category.ParentID()),
		"template_id":     category.TemplateID(),
		"code":            category.Code(),
		"name":            category.Name(),
		"sort_order":      category.SortOrder(),
		"created_at":      category.CreatedAt(),
		"updated_at":      updatedAt,
	}

	if err := r.dbFrom(ctx).WithContext(ctx).Table("catalog.product_categories").Create(payload).Error; err != nil {
		if isUniqueViolation(err) {
			return catalogerrors.ProductCategoryAlreadyExists()
		}
		if isForeignKeyViolation(err) {
			return catalogerrors.InvalidInput("Organization or parent category not found")
		}
		return err
	}
	return nil
}

func (r *CatalogRepo) UpdateProductCategory(ctx context.Context, category *catalogdomain.ProductCategory) error {
	if category == nil {
		return catalogerrors.InvalidInput("Product category is required")
	}
	updatedAt := category.CreatedAt()
	if category.UpdatedAt() != nil {
		updatedAt = *category.UpdatedAt()
	}
	payload := map[string]any{
		"parent_id":  nullableProductCategoryID(category.ParentID()),
		"code":       category.Code(),
		"name":       category.Name(),
		"sort_order": category.SortOrder(),
		"updated_at": updatedAt,
	}
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.product_categories").
		Where("organization_id = ? AND id = ? AND deleted_at IS NULL", category.OrganizationID().UUID(), category.ID().UUID()).
		Updates(payload).Error; err != nil {
		if isUniqueViolation(err) {
			return catalogerrors.ProductCategoryAlreadyExists()
		}
		if isForeignKeyViolation(err) {
			return catalogerrors.InvalidInput("Organization or parent category not found")
		}
		return err
	}
	return nil
}

func (r *CatalogRepo) DeleteProductCategory(ctx context.Context, organizationID orgdomain.OrganizationID, categoryID catalogdomain.ProductCategoryID, deletedAt time.Time) error {
	db := r.dbFrom(ctx).WithContext(ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("catalog.products").
			Where("organization_id = ? AND product_type_id = ? AND deleted_at IS NULL", organizationID.UUID(), categoryID.UUID()).
			Update("product_type_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Table("catalog.product_categories").
			Where("organization_id = ? AND parent_id = ? AND deleted_at IS NULL", organizationID.UUID(), categoryID.UUID()).
			Update("parent_id", nil).Error; err != nil {
			return err
		}
		return tx.Table("catalog.product_categories").
			Where("organization_id = ? AND id = ? AND deleted_at IS NULL", organizationID.UUID(), categoryID.UUID()).
			Updates(map[string]any{"deleted_at": deletedAt, "updated_at": deletedAt}).Error
	})
}

func (r *CatalogRepo) CreateProduct(ctx context.Context, product *catalogdomain.Product) error {
	if product == nil {
		return catalogerrors.InvalidInput("Product is required")
	}

	updatedAt := product.CreatedAt()
	if product.UpdatedAt() != nil {
		updatedAt = *product.UpdatedAt()
	}

	payload := map[string]any{
		"id":              product.ID().UUID(),
		"organization_id": product.OrganizationID().UUID(),
		"product_type_id": nullableProductCategoryID(product.CategoryID()),
		"name":            product.Name(),
		"description":     product.Description(),
		"sku":             product.SKU(),
		"price_amount":    product.PriceAmount(),
		"currency_code":   product.CurrencyCode(),
		"is_active":       product.IsActive(),
		"created_at":      product.CreatedAt(),
		"updated_at":      updatedAt,
	}

	if err := r.dbFrom(ctx).WithContext(ctx).Table("catalog.products").Create(payload).Error; err != nil {
		if isUniqueViolation(err) {
			return catalogerrors.ProductAlreadyExists()
		}
		if isForeignKeyViolation(err) {
			return catalogerrors.InvalidInput("Organization or product category not found")
		}
		return err
	}
	return nil
}

func (r *CatalogRepo) UpdateProduct(ctx context.Context, product *catalogdomain.Product) error {
	if product == nil {
		return catalogerrors.InvalidInput("Product is required")
	}
	updatedAt := product.CreatedAt()
	if product.UpdatedAt() != nil {
		updatedAt = *product.UpdatedAt()
	}
	payload := map[string]any{
		"product_type_id": nullableProductCategoryID(product.CategoryID()),
		"name":            product.Name(),
		"description":     product.Description(),
		"sku":             product.SKU(),
		"price_amount":    product.PriceAmount(),
		"currency_code":   product.CurrencyCode(),
		"is_active":       product.IsActive(),
		"updated_at":      updatedAt,
	}
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.products").
		Where("organization_id = ? AND id = ? AND deleted_at IS NULL", product.OrganizationID().UUID(), product.ID().UUID()).
		Updates(payload).Error; err != nil {
		if isUniqueViolation(err) {
			return catalogerrors.ProductAlreadyExists()
		}
		if isForeignKeyViolation(err) {
			return catalogerrors.InvalidInput("Organization or product category not found")
		}
		return err
	}
	return nil
}

func (r *CatalogRepo) DeleteProduct(ctx context.Context, organizationID orgdomain.OrganizationID, productID catalogdomain.ProductID, deletedAt time.Time) error {
	return r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.products").
		Where("organization_id = ? AND id = ? AND deleted_at IS NULL", organizationID.UUID(), productID.UUID()).
		Updates(map[string]any{"deleted_at": deletedAt, "updated_at": deletedAt}).Error
}

func (r *CatalogRepo) CreateStorageObject(ctx context.Context, object *ports.StorageObject) error {
	if object == nil {
		return catalogerrors.InvalidInput("Storage object is required")
	}

	payload := map[string]any{
		"id":              object.ID,
		"organization_id": object.OrganizationID.UUID(),
		"bucket":          object.Bucket,
		"object_key":      object.ObjectKey,
		"file_name":       object.FileName,
		"content_type":    object.ContentType,
		"size_bytes":      object.SizeBytes,
		"checksum_sha256": object.ChecksumSHA256,
		"created_at":      object.CreatedAt,
		"deleted_at":      object.DeletedAt,
	}

	if err := r.dbFrom(ctx).WithContext(ctx).Table("storage.objects").Create(payload).Error; err != nil {
		if isUniqueViolation(err) {
			return catalogerrors.InvalidInput("Storage object already exists")
		}
		if isForeignKeyViolation(err) {
			return catalogerrors.InvalidInput("Organization not found")
		}
		return err
	}
	return nil
}

func (r *CatalogRepo) CreateProductImportBatch(ctx context.Context, batch *ports.ProductImportBatch) error {
	if batch == nil {
		return catalogerrors.InvalidInput("Product import batch is required")
	}

	payload := map[string]any{
		"id":                    batch.ID,
		"organization_id":       batch.OrganizationID.UUID(),
		"source_object_id":      batch.SourceObjectID,
		"created_by_account_id": batch.CreatedByAccountID.UUID(),
		"status":                string(batch.Status),
		"total_rows":            batch.TotalRows,
		"processed_rows":        batch.ProcessedRows,
		"success_rows":          batch.SuccessRows,
		"error_rows":            batch.ErrorRows,
		"started_by":            batch.StartedBy,
		"started_at":            batch.StartedAt,
		"finished_at":           batch.FinishedAt,
		"created_at":            batch.CreatedAt,
		"updated_at":            batch.UpdatedAt,
		"mode":                  batch.Mode,
		"result_summary":        jsonbExpr(batch.ResultSummary),
	}

	return r.dbFrom(ctx).WithContext(ctx).Table("catalog.product_import_batches").Create(payload).Error
}

func (r *CatalogRepo) UpdateProductImportBatch(ctx context.Context, batch *ports.ProductImportBatch) error {
	if batch == nil {
		return catalogerrors.InvalidInput("Product import batch is required")
	}

	payload := map[string]any{
		"status":         string(batch.Status),
		"total_rows":     batch.TotalRows,
		"processed_rows": batch.ProcessedRows,
		"success_rows":   batch.SuccessRows,
		"error_rows":     batch.ErrorRows,
		"started_by":     batch.StartedBy,
		"started_at":     batch.StartedAt,
		"finished_at":    batch.FinishedAt,
		"updated_at":     batch.UpdatedAt,
		"mode":           batch.Mode,
		"result_summary": jsonbExpr(batch.ResultSummary),
	}

	return r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.product_import_batches").
		Where("organization_id = ? AND id = ?", batch.OrganizationID.UUID(), batch.ID).
		Updates(payload).Error
}

func (r *CatalogRepo) AddProductImportErrors(ctx context.Context, batchID uuid.UUID, items []ports.ProductImportErrorRecord) error {
	if len(items) == 0 {
		return nil
	}

	db := r.dbFrom(ctx).WithContext(ctx)
	for _, item := range items {
		payload := map[string]any{
			"id":         item.ID,
			"batch_id":   batchID,
			"row_no":     item.RowNo,
			"code":       item.Code,
			"message":    item.Message,
			"details":    jsonbExpr(item.Details),
			"created_at": item.CreatedAt,
		}
		if err := db.Table("catalog.product_import_errors").Create(payload).Error; err != nil {
			return err
		}
	}
	return nil
}

func nullableProductCategoryID(id *catalogdomain.ProductCategoryID) any {
	if id == nil {
		return nil
	}
	return id.UUID()
}

func jsonbExpr(value map[string]any) any {
	if value == nil {
		value = map[string]any{}
	}
	data, err := json.Marshal(value)
	if err != nil {
		data = []byte("{}")
	}
	return gorm.Expr("?::jsonb", string(data))
}
