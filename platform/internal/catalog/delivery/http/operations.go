package http

import "github.com/danielgtaylor/huma/v2"

var createProductCategoryOp = huma.Operation{
	OperationID: "create-product-category",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/product-categories",
	Tags:        []string{"Catalog"},
	Summary:     "Create product category",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listProductCategoriesOp = huma.Operation{
	OperationID: "list-product-categories",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/product-categories",
	Tags:        []string{"Catalog"},
	Summary:     "List product categories",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createProductOp = huma.Operation{
	OperationID: "create-product",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/products",
	Tags:        []string{"Catalog"},
	Summary:     "Create product",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listProductsOp = huma.Operation{
	OperationID: "list-products",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/products",
	Tags:        []string{"Catalog"},
	Summary:     "List products",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getProductByIDOp = huma.Operation{
	OperationID: "get-product",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/products/{product_id}",
	Tags:        []string{"Catalog"},
	Summary:     "Get product by id",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
