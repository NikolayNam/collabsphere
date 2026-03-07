package mapper

import (
	"net/http"

	"github.com/NikolayNam/collabsphere/internal/catalog/delivery/http/dto"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	"github.com/google/uuid"
)

func ToProductCategoryResponse(category *catalogdomain.ProductCategory, status int) *dto.ProductCategoryResponse {
	if category == nil {
		return nil
	}
	return &dto.ProductCategoryResponse{
		Status: status,
		Body: dto.ProductCategoryBody{
			ID:             category.ID().UUID(),
			OrganizationID: category.OrganizationID().UUID(),
			ParentID:       toCategoryUUIDPtr(category.ParentID()),
			Code:           category.Code(),
			Name:           category.Name(),
			SortOrder:      category.SortOrder(),
			CreatedAt:      category.CreatedAt(),
		},
	}
}

func ToProductCategoriesResponse(categories []catalogdomain.ProductCategory) *dto.ProductCategoriesResponse {
	resp := &dto.ProductCategoriesResponse{Status: http.StatusOK}
	resp.Body.Items = make([]dto.ProductCategoryBody, 0, len(categories))
	for _, category := range categories {
		resp.Body.Items = append(resp.Body.Items, dto.ProductCategoryBody{
			ID:             category.ID().UUID(),
			OrganizationID: category.OrganizationID().UUID(),
			ParentID:       toCategoryUUIDPtr(category.ParentID()),
			Code:           category.Code(),
			Name:           category.Name(),
			SortOrder:      category.SortOrder(),
			CreatedAt:      category.CreatedAt(),
		})
	}
	return resp
}

func ToProductResponse(product *catalogdomain.Product, status int) *dto.ProductResponse {
	if product == nil {
		return nil
	}
	return &dto.ProductResponse{
		Status: status,
		Body: dto.ProductBody{
			ID:             product.ID().UUID(),
			OrganizationID: product.OrganizationID().UUID(),
			CategoryID:     toCategoryUUIDPtr(product.CategoryID()),
			Name:           product.Name(),
			Description:    product.Description(),
			SKU:            product.SKU(),
			PriceAmount:    product.PriceAmount(),
			CurrencyCode:   product.CurrencyCode(),
			IsActive:       product.IsActive(),
			CreatedAt:      product.CreatedAt(),
		},
	}
}

func ToProductsResponse(products []catalogdomain.Product) *dto.ProductsResponse {
	resp := &dto.ProductsResponse{Status: http.StatusOK}
	resp.Body.Items = make([]dto.ProductBody, 0, len(products))
	for _, product := range products {
		resp.Body.Items = append(resp.Body.Items, dto.ProductBody{
			ID:             product.ID().UUID(),
			OrganizationID: product.OrganizationID().UUID(),
			CategoryID:     toCategoryUUIDPtr(product.CategoryID()),
			Name:           product.Name(),
			Description:    product.Description(),
			SKU:            product.SKU(),
			PriceAmount:    product.PriceAmount(),
			CurrencyCode:   product.CurrencyCode(),
			IsActive:       product.IsActive(),
			CreatedAt:      product.CreatedAt(),
		})
	}
	return resp
}

func toCategoryUUIDPtr(id *catalogdomain.ProductCategoryID) *uuid.UUID {
	if id == nil {
		return nil
	}
	value := id.UUID()
	return &value
}
