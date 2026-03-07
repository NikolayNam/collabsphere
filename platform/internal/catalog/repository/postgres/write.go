package postgres

import (
	"context"

	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
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

func nullableProductCategoryID(id *catalogdomain.ProductCategoryID) any {
	if id == nil {
		return nil
	}
	return id.UUID()
}
