package postgres

import (
	"context"
	"errors"
	"time"

	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

func (r *CatalogRepo) ListProductCategories(ctx context.Context, organizationID orgdomain.OrganizationID) ([]catalogdomain.ProductCategory, error) {
	var rows []catalogProductCategoryRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.product_categories").
		Select("id", "organization_id", "parent_id", "template_id", "code", "name", "sort_order", "created_at", "updated_at").
		Where("organization_id = ? AND deleted_at IS NULL", organizationID.UUID()).
		Order("sort_order ASC, name ASC, id ASC").
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

func (r *CatalogRepo) ListProducts(ctx context.Context, organizationID orgdomain.OrganizationID) ([]catalogdomain.Product, error) {
	var rows []catalogProductRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("catalog.products").
		Select("id", "organization_id", "product_type_id", "name", "description", "sku", "price_amount::text AS price_amount", "currency_code", "is_active", "created_at", "updated_at").
		Where("organization_id = ? AND deleted_at IS NULL", organizationID.UUID()).
		Order("created_at DESC, id DESC").
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
