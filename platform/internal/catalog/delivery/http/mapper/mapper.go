package mapper

import (
	"net/http"

	catalogapp "github.com/NikolayNam/collabsphere/internal/catalog/application"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
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

func ToProductImportResponse(view *catalogapp.ProductImportView, status int) *dto.ProductImportResponse {
	if view == nil || view.Batch == nil {
		return nil
	}
	return &dto.ProductImportResponse{
		Status: status,
		Body: dto.ProductImportBatchBody{
			ID:                 view.Batch.ID,
			OrganizationID:     view.Batch.OrganizationID.UUID(),
			SourceObjectID:     view.Batch.SourceObjectID,
			CreatedByAccountID: view.Batch.CreatedByAccountID.UUID(),
			Status:             string(view.Batch.Status),
			TotalRows:          view.Batch.TotalRows,
			ProcessedRows:      view.Batch.ProcessedRows,
			SuccessRows:        view.Batch.SuccessRows,
			ErrorRows:          view.Batch.ErrorRows,
			StartedBy:          view.Batch.StartedBy,
			StartedAt:          view.Batch.StartedAt,
			FinishedAt:         view.Batch.FinishedAt,
			CreatedAt:          view.Batch.CreatedAt,
			UpdatedAt:          view.Batch.UpdatedAt,
			Mode:               view.Batch.Mode,
			ResultSummary:      copyMap(view.Batch.ResultSummary),
			Errors:             toImportErrors(view.Errors),
		},
	}
}

func toImportErrors(items []ports.ProductImportErrorRecord) []dto.ProductImportErrorBody {
	out := make([]dto.ProductImportErrorBody, 0, len(items))
	for _, item := range items {
		out = append(out, dto.ProductImportErrorBody{
			ID:        item.ID,
			RowNo:     item.RowNo,
			Code:      item.Code,
			Message:   item.Message,
			Details:   copyMap(item.Details),
			CreatedAt: item.CreatedAt,
		})
	}
	return out
}

func toCategoryUUIDPtr(id *catalogdomain.ProductCategoryID) *uuid.UUID {
	if id == nil {
		return nil
	}
	value := id.UUID()
	return &value
}

func copyMap(value map[string]any) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(value))
	for key, item := range value {
		out[key] = item
	}
	return out
}
