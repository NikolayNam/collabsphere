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

var updateProductCategoryOp = huma.Operation{
	OperationID: "update-product-category",
	Method:      "PATCH",
	Path:        "/organizations/{organization_id}/product-categories/{category_id}",
	Tags:        []string{"Catalog"},
	Summary:     "Update product category",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var deleteProductCategoryOp = huma.Operation{
	OperationID: "delete-product-category",
	Method:      "DELETE",
	Path:        "/organizations/{organization_id}/product-categories/{category_id}",
	Tags:        []string{"Catalog"},
	Summary:     "Delete product category",
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

var updateProductOp = huma.Operation{
	OperationID: "update-product",
	Method:      "PATCH",
	Path:        "/organizations/{organization_id}/products/{product_id}",
	Tags:        []string{"Catalog"},
	Summary:     "Update product",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var deleteProductOp = huma.Operation{
	OperationID: "delete-product",
	Method:      "DELETE",
	Path:        "/organizations/{organization_id}/products/{product_id}",
	Tags:        []string{"Catalog"},
	Summary:     "Delete product",
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

var createProductImportUploadOp = huma.Operation{
	OperationID: "create-product-import-upload",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/product-imports/uploads",
	Tags:        []string{"Catalog"},
	Summary:     "Create presigned upload for product import file",
	Description: "This endpoint does not accept multipart file content. Send JSON metadata to receive a presigned upload URL. Then upload the raw file bytes with HTTP PUT to body.uploadUrl. After the upload succeeds, call POST /api/v1/organizations/{organization_id}/product-imports with sourceObjectId = body.objectId.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var runProductImportOp = huma.Operation{
	OperationID: "run-product-import",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/product-imports",
	Tags:        []string{"Catalog"},
	Summary:     "Process product and category import file",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getProductImportOp = huma.Operation{
	OperationID: "get-product-import",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/product-imports/{batch_id}",
	Tags:        []string{"Catalog"},
	Summary:     "Get product import batch",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
