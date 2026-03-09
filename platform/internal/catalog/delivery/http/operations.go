package http

import "github.com/danielgtaylor/huma/v2"

var createProductCategoryOp = huma.Operation{
	OperationID: "create-product-category",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/product-categories",
	Tags:        []string{"Organizations / Catalog"},
	Summary:     "Create product category",
	Description: "Creates an organization-scoped product category. Categories belong to a single organization and are not shared globally between tenants.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var updateProductCategoryOp = huma.Operation{
	OperationID: "update-product-category",
	Method:      "PATCH",
	Path:        "/organizations/{organization_id}/product-categories/{category_id}",
	Tags:        []string{"Organizations / Catalog"},
	Summary:     "Update product category",
	Description: "Updates mutable fields of an organization-scoped product category, such as title, parent category, sort order, or visibility state.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var deleteProductCategoryOp = huma.Operation{
	OperationID: "delete-product-category",
	Method:      "DELETE",
	Path:        "/organizations/{organization_id}/product-categories/{category_id}",
	Tags:        []string{"Organizations / Catalog"},
	Summary:     "Delete product category",
	Description: "Removes a product category from the organization catalog when domain invariants allow it.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listProductCategoriesOp = huma.Operation{
	OperationID: "list-product-categories",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/product-categories",
	Tags:        []string{"Organizations / Catalog"},
	Summary:     "List product categories",
	Description: "Returns the product categories available inside a single organization catalog.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createProductOp = huma.Operation{
	OperationID: "create-product",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/products",
	Tags:        []string{"Organizations / Catalog"},
	Summary:     "Create product",
	Description: "Creates a product in the organization catalog and links it to an organization-scoped category.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var updateProductOp = huma.Operation{
	OperationID: "update-product",
	Method:      "PATCH",
	Path:        "/organizations/{organization_id}/products/{product_id}",
	Tags:        []string{"Organizations / Catalog"},
	Summary:     "Update product",
	Description: "Updates mutable product fields such as name, SKU, description, pricing, and category assignment.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var deleteProductOp = huma.Operation{
	OperationID: "delete-product",
	Method:      "DELETE",
	Path:        "/organizations/{organization_id}/products/{product_id}",
	Tags:        []string{"Organizations / Catalog"},
	Summary:     "Delete product",
	Description: "Removes a product from the organization catalog.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listProductsOp = huma.Operation{
	OperationID: "list-products",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/products",
	Tags:        []string{"Organizations / Catalog"},
	Summary:     "List products",
	Description: "Returns the products that belong to the organization catalog.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getProductByIDOp = huma.Operation{
	OperationID: "get-product",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/products/{product_id}",
	Tags:        []string{"Organizations / Catalog"},
	Summary:     "Get product by id",
	Description: "Returns a single product from the organization catalog by product id.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var uploadProductImportOp = huma.Operation{
	OperationID: "upload-product-import",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/product-imports/upload",
	Tags:        []string{"Organizations / Catalog"},
	Summary:     "Upload and process product and category import file directly",
	Description: "Single-step product import using multipart/form-data. Send the CSV file in the `file` field. The backend uploads the object to S3-compatible storage and immediately runs the import in mode `upsert`.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var runProductImportOp = huma.Operation{
	OperationID: "run-product-import",
	Method:      "POST",
	Path:        "/organizations/{organization_id}/product-imports",
	Tags:        []string{"Organizations / Catalog"},
	Summary:     "Process product and category import file",
	Description: "Processes a previously registered import source object and creates or updates categories and products in the target organization catalog.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getProductImportOp = huma.Operation{
	OperationID: "get-product-import",
	Method:      "GET",
	Path:        "/organizations/{organization_id}/product-imports/{batch_id}",
	Tags:        []string{"Organizations / Catalog"},
	Summary:     "Get product import batch",
	Description: "Returns the current state of a product import batch, including counters, source object metadata, and validation or import errors.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
